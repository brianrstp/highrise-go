package highrise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const webapiBaseURL = "https://webapi.highrise.game"

type WebAPI struct {
	client  *http.Client
	baseURL string
}

func NewWebAPI() *WebAPI {
	return &WebAPI{
		client:  &http.Client{},
		baseURL: webapiBaseURL,
	}
}

func (w *WebAPI) getJSON(path string, query url.Values, dest any) error {
	u, err := url.Parse(w.baseURL + path)
	if err != nil {
		return err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	resp, err := w.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("webapi request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("webapi read failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webapi returned status %d: %s", resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("webapi decode failed: %w", err)
	}
	return nil
}

type UserData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type UserResponse struct {
	User UserData `json:"user"`
}

func (w *WebAPI) GetUser(ctx context.Context, userID string) (*UserData, error) {
	var resp UserResponse
	if err := w.getJSON("/users/"+userID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

type UsersListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Username    string
}

type UsersListResponse struct {
	Users []UserData `json:"users"`
}

func (w *WebAPI) GetUsers(ctx context.Context, params UsersListParams) ([]UserData, error) {
	q := url.Values{}
	if params.StartsAfter != "" {
		q.Set("starts_after", params.StartsAfter)
	}
	if params.EndsBefore != "" {
		q.Set("ends_before", params.EndsBefore)
	}
	if params.SortOrder != "" {
		q.Set("sort_order", params.SortOrder)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Username != "" {
		q.Set("username", params.Username)
	}
	var resp UsersListResponse
	if err := w.getJSON("/users", q, &resp); err != nil {
		return nil, err
	}
	return resp.Users, nil
}

type RoomData struct {
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name"`
	OwnerID  string `json:"owner_id"`
}

type RoomResponse struct {
	Room RoomData `json:"room"`
}

func (w *WebAPI) GetRoom(ctx context.Context, roomID string) (*RoomData, error) {
	var resp RoomResponse
	if err := w.getJSON("/rooms/"+roomID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Room, nil
}

type RoomsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	RoomName    string
	OwnerID     string
}

type RoomsListResponse struct {
	Rooms []RoomData `json:"rooms"`
}

func (w *WebAPI) GetRooms(ctx context.Context, params RoomsListParams) ([]RoomData, error) {
	q := url.Values{}
	if params.StartsAfter != "" {
		q.Set("starts_after", params.StartsAfter)
	}
	if params.EndsBefore != "" {
		q.Set("ends_before", params.EndsBefore)
	}
	if params.SortOrder != "" {
		q.Set("sort_order", params.SortOrder)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.RoomName != "" {
		q.Set("room_name", params.RoomName)
	}
	if params.OwnerID != "" {
		q.Set("owner_id", params.OwnerID)
	}
	var resp RoomsListResponse
	if err := w.getJSON("/rooms", q, &resp); err != nil {
		return nil, err
	}
	return resp.Rooms, nil
}

type ItemData struct {
	ItemID   string  `json:"item_id"`
	ItemName string  `json:"item_name"`
	Type     string  `json:"type"`
	Rarity   *string `json:"rarity,omitempty"`
	Category *string `json:"category,omitempty"`
}

type ItemResponse struct {
	Item ItemData `json:"item"`
}

func (w *WebAPI) GetItem(ctx context.Context, itemID string) (*ItemData, error) {
	var resp ItemResponse
	if err := w.getJSON("/items/"+itemID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

type ItemsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Rarity      string
	ItemName    string
	Category    string
}

type ItemsListResponse struct {
	Items []ItemData `json:"items"`
}

func (w *WebAPI) GetItems(ctx context.Context, params ItemsListParams) ([]ItemData, error) {
	q := url.Values{}
	if params.StartsAfter != "" {
		q.Set("starts_after", params.StartsAfter)
	}
	if params.EndsBefore != "" {
		q.Set("ends_before", params.EndsBefore)
	}
	if params.SortOrder != "" {
		q.Set("sort_order", params.SortOrder)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Rarity != "" {
		q.Set("rarity", params.Rarity)
	}
	if params.ItemName != "" {
		q.Set("item_name", params.ItemName)
	}
	if params.Category != "" {
		q.Set("category", params.Category)
	}
	var resp ItemsListResponse
	if err := w.getJSON("/items", q, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

type GrabData struct {
	GrabID string `json:"grab_id"`
	Title  string `json:"title"`
}

type GrabResponse struct {
	Grab GrabData `json:"grab"`
}

func (w *WebAPI) GetGrab(ctx context.Context, grabID string) (*GrabData, error) {
	var resp GrabResponse
	if err := w.getJSON("/grabs/"+grabID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Grab, nil
}

type GrabsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Title       string
}

type GrabsListResponse struct {
	Grabs []GrabData `json:"grabs"`
}

func (w *WebAPI) GetGrabs(ctx context.Context, params GrabsListParams) ([]GrabData, error) {
	q := url.Values{}
	if params.StartsAfter != "" {
		q.Set("starts_after", params.StartsAfter)
	}
	if params.EndsBefore != "" {
		q.Set("ends_before", params.EndsBefore)
	}
	if params.SortOrder != "" {
		q.Set("sort_order", params.SortOrder)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Title != "" {
		q.Set("title", params.Title)
	}
	var resp GrabsListResponse
	if err := w.getJSON("/grabs", q, &resp); err != nil {
		return nil, err
	}
	return resp.Grabs, nil
}

type PostData struct {
	PostID   string `json:"post_id"`
	AuthorID string `json:"author_id"`
	Content  string `json:"content"`
}

type PostResponse struct {
	Post PostData `json:"post"`
}

func (w *WebAPI) GetPost(ctx context.Context, postID string) (*PostData, error) {
	var resp PostResponse
	if err := w.getJSON("/posts/"+postID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Post, nil
}

type PostsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	AuthorID    string
}

type PostsListResponse struct {
	Posts []PostData `json:"posts"`
}

func (w *WebAPI) GetPosts(ctx context.Context, params PostsListParams) ([]PostData, error) {
	q := url.Values{}
	if params.StartsAfter != "" {
		q.Set("starts_after", params.StartsAfter)
	}
	if params.EndsBefore != "" {
		q.Set("ends_before", params.EndsBefore)
	}
	if params.SortOrder != "" {
		q.Set("sort_order", params.SortOrder)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.AuthorID != "" {
		q.Set("author_id", params.AuthorID)
	}
	var resp PostsListResponse
	if err := w.getJSON("/posts", q, &resp); err != nil {
		return nil, err
	}
	return resp.Posts, nil
}
