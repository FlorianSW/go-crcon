package crcon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type client struct {
	hc http.Client

	baseUrl string
	creds   Credentials
}

func NewClient(hc http.Client, baseUrl string, creds Credentials) *client {
	return &client{hc: hc, baseUrl: baseUrl, creds: creds}
}

func (c *client) Matches(ctx context.Context) ([]Match, error) {
	u, err := url.JoinPath(c.baseUrl, "/api/get_recent_logs")
	if err != nil {
		return nil, err
	}
	req, err := json.Marshal(getRecentLogsRequest{
		End:             5000,
		FilterActions:   []action{ActionMatchEnded, ActionMatchStart},
		InclusiveFilter: true,
	})
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("POST", u, bytes.NewReader(req))
	if err != nil {
		return nil, err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	result, err := asResponse[getRecentLogsResponse](res)
	if err != nil {
		return nil, err
	}
	slices.Reverse(result.Logs)
	var matches []Match
	for _, l := range result.Logs {
		t := time.Time(l.EventTime)

		if l.Action == ActionMatchStart && len(matches) > 0 && matches[len(matches)-1].End == nil {
			matches[len(matches)-1].End = &t
		}
		if l.Action == ActionMatchStart {
			matches = append(matches, Match{
				Start: &t,
				End:   nil,
				Map:   l.SubContent,
				Score: Score{},
			})
		}
		if l.Action == ActionMatchEnded && len(matches) > 0 {
			m, score := parseMapAndResult(l.SubContent)
			matches[len(matches)-1].End = &t
			matches[len(matches)-1].Map = m
			matches[len(matches)-1].Score = score
		}
	}
	return matches, nil
}

func parseMapAndResult(c string) (string, Score) {
	// `ELSENBORN RIDGE Skirmish` ALLIED (0 - 0) AXIS
	p := strings.Split(c, "`")
	s := strings.ReplaceAll(p[2], "ALLIED (", "")
	s = strings.ReplaceAll(s, ") AXIS", "")
	score := strings.Split(s, " - ")

	allied, _ := strconv.Atoi(strings.TrimSpace(score[0]))
	axis, _ := strconv.Atoi(strings.TrimSpace(score[1]))
	return p[1], Score{Allied: allied, Axis: axis}
}

func (c *client) SwitchMap(ctx context.Context, id string) error {
	u, err := url.JoinPath(c.baseUrl, "/api/set_map")
	if err != nil {
		return err
	}
	req, err := json.Marshal(setMapRequest{
		MapId: id,
	})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("POST", u, bytes.NewReader(req))
	if err != nil {
		return err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil
}

func (c *client) MapRotation(ctx context.Context) (MapRotation, error) {
	u, err := url.JoinPath(c.baseUrl, "/api/get_map_rotation")
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	result, err := asResponse[getMapRotationResponse](res)
	if err != nil {
		return nil, err
	}
	return result.toMapRotation(), nil
}

func (c *client) MessagePlayer(ctx context.Context, playerId, message string) error {
	u, err := url.JoinPath(c.baseUrl, "/api/message_player")
	if err != nil {
		return err
	}
	req, err := json.Marshal(messagePlayerRequest{
		PlayerId: playerId,
		Message:  message,
	})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("POST", u, bytes.NewReader(req))
	if err != nil {
		return err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return nil
}

func (c *client) GameState(ctx context.Context) (GameState, error) {
	u, err := url.JoinPath(c.baseUrl, "/api/get_gamestate")
	if err != nil {
		return GameState{}, err
	}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return GameState{}, err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return GameState{}, err
	}
	if res.StatusCode == http.StatusForbidden {
		return GameState{}, ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return GameState{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	result, err := asResponse[getGameStateResponse](res)
	if err != nil {
		return GameState{}, err
	}
	return result.toGameState(), nil
}

func (c *client) PlayerIds(ctx context.Context) ([]string, error) {
	u, err := url.JoinPath(c.baseUrl, "/api/get_playerids")
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("GET", u+"?as_dict=true", nil)
	if err != nil {
		return nil, err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	result, err := asResponse[map[string]string](res)
	if err != nil {
		return nil, err
	}
	return slices.Collect(maps.Values(result)), nil
}

func (c *client) OwnPermissions(ctx context.Context) (OwnPermissions, error) {
	u, err := url.JoinPath(c.baseUrl, "/api/get_own_user_permissions")
	if err != nil {
		return OwnPermissions{}, err
	}
	r, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return OwnPermissions{}, err
	}
	r = r.WithContext(ctx)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.ApiKey))

	res, err := c.hc.Do(r)
	if err != nil {
		return OwnPermissions{}, err
	}
	if res.StatusCode == http.StatusForbidden {
		return OwnPermissions{}, ErrForbidden
	}
	if res.StatusCode != http.StatusOK {
		return OwnPermissions{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	result, err := asResponse[getOwnPermissions](res)
	if err != nil {
		return OwnPermissions{}, err
	}
	return result.toOwnPermissions(), nil
}
