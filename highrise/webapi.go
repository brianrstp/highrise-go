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

// WebAPI provides a REST client for the Highrise WebAPI
type WebAPI struct {
	client  *http.Client
	baseURL string
}

// NewWebAPI creates a new WebAPI client
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

// UserData contains basic user information from the WebAPI
type UserData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// UserResponse wraps a single user response
type UserResponse struct {
	User UserData `json:"user"`
}

// GetUser gets a user by ID
func (w *WebAPI) GetUser(ctx context.Context, userID string) (*UserData, error) {
	var resp UserResponse
	if err := w.getJSON("/users/"+userID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// UsersListParams contains optional parameters for listing users
type UsersListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Username    string
}

// UsersListResponse contains a paginated list of users
type UsersListResponse struct {
	Users     []UserData `json:"users"`
	Total     int        `json:"total"`
	FirstID   string     `json:"first_id"`
	LastID    string     `json:"last_id"`
}

// GetUsers lists users with optional filters
func (w *WebAPI) GetUsers(ctx context.Context, params UsersListParams) (*UsersListResponse, error) {
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
	return &resp, nil
}

// RoomData contains room information from the WebAPI
type RoomData struct {
	RoomID       string   `json:"room_id"`
	DispName     string   `json:"disp_name"`
	Description  *string  `json:"description"`
	Category     string   `json:"category"`
	CreatedAt    string   `json:"created_at"`
	AccessPolicy string   `json:"access_policy"`
	OwnerID      *string  `json:"owner_id"`
	Locale       []string `json:"locale"`
	IsHomeRoom   bool     `json:"is_home_room"`
	DesignerIDs  []string `json:"designer_ids"`
	ModeratorIDs []string `json:"moderator_ids"`
}

// RoomResponse wraps a single room response
type RoomResponse struct {
	Room RoomData `json:"room"`
}

// GetRoom gets a room by ID
func (w *WebAPI) GetRoom(ctx context.Context, roomID string) (*RoomData, error) {
	var resp RoomResponse
	if err := w.getJSON("/rooms/"+roomID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Room, nil
}

// RoomsListParams contains optional parameters for listing rooms
type RoomsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	RoomName    string
	OwnerID     string
}

// RoomsListResponse contains a paginated list of rooms
type RoomsListResponse struct {
	Rooms   []RoomData `json:"rooms"`
	Total   int        `json:"total"`
	FirstID string     `json:"first_id"`
	LastID  string     `json:"last_id"`
}

// GetRooms lists rooms with optional filters
func (w *WebAPI) GetRooms(ctx context.Context, params RoomsListParams) (*RoomsListResponse, error) {
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
	return &resp, nil
}

// ItemData contains item information from the WebAPI
type ItemData struct {
	ItemID              string   `json:"item_id"`
	ItemName            string   `json:"item_name"`
	Type                string   `json:"type"`
	Category            string   `json:"category"`
	Rarity              string   `json:"rarity"`
	CreatedAt           string   `json:"created_at"`
	IsPurchasable       bool     `json:"is_purchasable"`
	IsTradable          bool     `json:"is_tradable"`
	PopsSalePrice       int      `json:"pops_sale_price"`
	GemsSalePrice       *int     `json:"gems_sale_price,omitempty"`
	DescriptionKey      *string  `json:"description_key,omitempty"`
	Keywords            []string `json:"keywords,omitempty"`
	ImageURL            *string  `json:"image_url,omitempty"`
	IconURL             *string  `json:"icon_url,omitempty"`
}

// ItemResponse wraps a single item response
type ItemResponse struct {
	Item ItemData `json:"item"`
}

// GetItem gets an item by ID
func (w *WebAPI) GetItem(ctx context.Context, itemID string) (*ItemData, error) {
	var resp ItemResponse
	if err := w.getJSON("/items/"+itemID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// ItemsListParams contains optional parameters for listing items
type ItemsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Rarity      string
	ItemName    string
	Category    string
}

// ItemsListResponse contains a paginated list of items
type ItemsListResponse struct {
	Items   []ItemData `json:"items"`
	Total   int        `json:"total"`
	FirstID string     `json:"first_id"`
	LastID  string     `json:"last_id"`
}

// GetItems lists items with optional filters
func (w *WebAPI) GetItems(ctx context.Context, params ItemsListParams) (*ItemsListResponse, error) {
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
	return &resp, nil
}

// GrabData contains grab bag information from the WebAPI
type GrabData struct {
	GrabID       string   `json:"grab_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ItemIDs      []string `json:"item_ids"`
	SalePrice    int      `json:"sale_price"`
	SalePriceGems *int    `json:"sale_price_gems,omitempty"`
	StartsAt     string   `json:"starts_at"`
	EndsAt       string   `json:"ends_at"`
	ImageURL     *string  `json:"image_url,omitempty"`
}

// GrabResponse wraps a single grab response
type GrabResponse struct {
	Grab GrabData `json:"grab"`
}

// GetGrab gets a grab bag by ID
func (w *WebAPI) GetGrab(ctx context.Context, grabID string) (*GrabData, error) {
	var resp GrabResponse
	if err := w.getJSON("/grabs/"+grabID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Grab, nil
}

// GrabsListParams contains optional parameters for listing grab bags
type GrabsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	Title       string
}

// GrabsListResponse contains a paginated list of grab bags
type GrabsListResponse struct {
	Grabs   []GrabData `json:"grabs"`
	Total   int        `json:"total"`
	FirstID string     `json:"first_id"`
	LastID  string     `json:"last_id"`
}

// GetGrabs lists grab bags with optional filters
func (w *WebAPI) GetGrabs(ctx context.Context, params GrabsListParams) (*GrabsListResponse, error) {
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
	return &resp, nil
}

// PostData contains post information from the WebAPI
type PostData struct {
	PostID         string      `json:"post_id"`
	AuthorID       string      `json:"author_id"`
	CreatedAt      string      `json:"created_at"`
	FileKey        *string     `json:"file_key"`
	Type           string      `json:"type"`
	Visibility     string      `json:"visibility"`
	NumComments    int         `json:"num_comments"`
	NumLikes       int         `json:"num_likes"`
	NumReposts     int         `json:"num_reposts"`
	Caption        *string     `json:"caption"`
	FeaturedUserIDs []string   `json:"featured_user_ids"`
}

// PostResponse wraps a single post response
type PostResponse struct {
	Post PostData `json:"post"`
}

// GetPost gets a post by ID
func (w *WebAPI) GetPost(ctx context.Context, postID string) (*PostData, error) {
	var resp PostResponse
	if err := w.getJSON("/posts/"+postID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Post, nil
}

// PostsListParams contains optional parameters for listing posts
type PostsListParams struct {
	StartsAfter string
	EndsBefore  string
	SortOrder   string
	Limit       int
	AuthorID    string
}

// PostsListResponse contains a paginated list of posts
type PostsListResponse struct {
	Posts   []PostData `json:"posts"`
	Total   int        `json:"total"`
	FirstID string     `json:"first_id"`
	LastID  string     `json:"last_id"`
}

// GetPosts lists posts with optional filters
func (w *WebAPI) GetPosts(ctx context.Context, params PostsListParams) (*PostsListResponse, error) {
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
	return &resp, nil
}
