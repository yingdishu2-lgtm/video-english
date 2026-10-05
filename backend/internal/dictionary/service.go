package dictionary

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.gdys01.cn/gdys01/English-media-learner/backend/ent"
	"git.gdys01.cn/gdys01/English-media-learner/backend/ent/wordpronunciation"
)

const defaultEndpoint = "https://api.dictionaryapi.dev/api/v2/entries/en/"

var englishWord = regexp.MustCompile(`^[A-Za-z]+(?:['’][A-Za-z]+)*$`)

type Service struct {
	client   *ent.Client
	http     *http.Client
	endpoint string
}

type Pronunciation struct {
	Word     string `json:"word"`
	Phonetic string `json:"phonetic,omitempty"`
	AudioURL string `json:"audioUrl,omitempty"`
	Source   string `json:"source"`
	Found    bool   `json:"found"`
	Cached   bool   `json:"cached"`
}

type Option func(*Service)

func WithEndpoint(endpoint string) Option {
	return func(service *Service) { service.endpoint = endpoint }
}

func WithHTTPClient(client *http.Client) Option {
	return func(service *Service) { service.http = client }
}

func NewService(client *ent.Client, options ...Option) *Service {
	service := &Service{client: client, endpoint: defaultEndpoint, http: &http.Client{Timeout: 6 * time.Second}}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) Lookup(ctx context.Context, input string) (Pronunciation, error) {
	word := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(input, "’", "'")))
	if !englishWord.MatchString(word) || len(word) > 80 {
		return Pronunciation{}, fmt.Errorf("invalid English word")
	}
	if cached, err := s.client.WordPronunciation.Query().Where(wordpronunciation.WordEQ(word)).Only(ctx); err == nil {
		return Pronunciation{Word: word, Phonetic: cached.Phonetic, AudioURL: cached.AudioURL,
			Source: cached.Source, Found: cached.Found, Cached: true}, nil
	} else if !ent.IsNotFound(err) {
		return Pronunciation{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint+url.PathEscape(word), nil)
	if err != nil {
		return Pronunciation{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := s.http.Do(request)
	if err != nil {
		return Pronunciation{}, fmt.Errorf("dictionary request: %w", err)
	}
	defer response.Body.Close()
	result := Pronunciation{Word: word, Source: "Free Dictionary API"}
	if response.StatusCode == http.StatusNotFound {
		if err := s.save(ctx, result); err != nil {
			return Pronunciation{}, err
		}
		return result, nil
	}
	if response.StatusCode != http.StatusOK {
		return Pronunciation{}, fmt.Errorf("dictionary service returned %s", response.Status)
	}
	var entries []struct {
		Phonetic  string `json:"phonetic"`
		Phonetics []struct {
			Text  string `json:"text"`
			Audio string `json:"audio"`
		} `json:"phonetics"`
	}
	if err := json.NewDecoder(response.Body).Decode(&entries); err != nil {
		return Pronunciation{}, fmt.Errorf("decode dictionary response: %w", err)
	}
	for _, entry := range entries {
		if result.Phonetic == "" {
			result.Phonetic = strings.TrimSpace(entry.Phonetic)
		}
		for _, phonetic := range entry.Phonetics {
			if result.Phonetic == "" && strings.TrimSpace(phonetic.Text) != "" {
				result.Phonetic = strings.TrimSpace(phonetic.Text)
			}
			if result.AudioURL == "" && strings.TrimSpace(phonetic.Audio) != "" {
				result.AudioURL = normalizeAudioURL(phonetic.Audio)
			}
		}
	}
	result.Found = result.Phonetic != "" || result.AudioURL != ""
	if err := s.save(ctx, result); err != nil {
		return Pronunciation{}, err
	}
	return result, nil
}

func (s *Service) save(ctx context.Context, result Pronunciation) error {
	_, err := s.client.WordPronunciation.Create().SetWord(result.Word).SetPhonetic(result.Phonetic).
		SetAudioURL(result.AudioURL).SetSource(result.Source).SetFound(result.Found).Save(ctx)
	if err != nil && ent.IsConstraintError(err) {
		return nil
	}
	return err
}

func normalizeAudioURL(value string) string {
	if strings.HasPrefix(value, "//") {
		return "https:" + value
	}
	return value
}
