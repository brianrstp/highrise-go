package highrise

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newWebAPITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]string{
				"user_id":  "u1",
				"username": "alice",
			},
		})
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"users": []map[string]string{
				{"user_id": "u1", "username": "alice"},
			},
			"total":    1,
			"first_id": "u1",
			"last_id":  "u1",
		})
	})

	mux.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"room": map[string]any{
				"room_id":   "r1",
				"disp_name": "Test Room",
				"category":  "hangout",
			},
		})
	})

	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"rooms": []map[string]any{
				{"room_id": "r1", "disp_name": "Test Room", "category": "hangout"},
			},
			"total":    1,
			"first_id": "r1",
			"last_id":  "r1",
		})
	})

	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"item": map[string]any{
				"item_id":   "i1",
				"item_name": "Test Item",
				"type":      "shirt",
				"category":  "top",
				"rarity":    "common",
			},
		})
	})

	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"item_id": "i1", "item_name": "Test Item", "type": "shirt", "category": "top", "rarity": "common"},
			},
			"total":    1,
			"first_id": "i1",
			"last_id":  "i1",
		})
	})

	mux.HandleFunc("/grabs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"grab": map[string]any{
				"grab_id": "g1",
				"title":   "Test Grab",
			},
		})
	})

	mux.HandleFunc("/grabs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"grabs": []map[string]any{
				{"grab_id": "g1", "title": "Test Grab"},
			},
			"total":    1,
			"first_id": "g1",
			"last_id":  "g1",
		})
	})

	mux.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"post": map[string]any{
				"post_id":   "p1",
				"author_id": "u1",
				"type":      "text",
				"visibility": "public",
			},
		})
	})

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"posts": []map[string]any{
				{"post_id": "p1", "author_id": "u1", "type": "text", "visibility": "public"},
			},
			"total":    1,
			"first_id": "p1",
			"last_id":  "p1",
		})
	})

	return httptest.NewServer(mux)
}

func TestWebAPI_GetUser(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	user, err := w.GetUser(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if user.UserID != "u1" || user.Username != "alice" {
		t.Fatal("unexpected user")
	}
}

func TestWebAPI_GetUsers(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	resp, err := w.GetUsers(context.Background(), UsersListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Users) != 1 || resp.Total != 1 {
		t.Fatal("unexpected users response")
	}
}

func TestWebAPI_GetRoom(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	room, err := w.GetRoom(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if room.RoomID != "r1" || room.DispName != "Test Room" {
		t.Fatal("unexpected room")
	}
}

func TestWebAPI_GetRooms(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	resp, err := w.GetRooms(context.Background(), RoomsListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Rooms) != 1 || resp.Total != 1 {
		t.Fatal("unexpected rooms response")
	}
}

func TestWebAPI_GetItem(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	item, err := w.GetItem(context.Background(), "i1")
	if err != nil {
		t.Fatal(err)
	}
	if item.ItemID != "i1" || item.ItemName != "Test Item" {
		t.Fatal("unexpected item")
	}
}

func TestWebAPI_GetItems(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	resp, err := w.GetItems(context.Background(), ItemsListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Items) != 1 || resp.Total != 1 {
		t.Fatal("unexpected items response")
	}
}

func TestWebAPI_GetGrab(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	grab, err := w.GetGrab(context.Background(), "g1")
	if err != nil {
		t.Fatal(err)
	}
	if grab.GrabID != "g1" || grab.Title != "Test Grab" {
		t.Fatal("unexpected grab")
	}
}

func TestWebAPI_GetGrabs(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	resp, err := w.GetGrabs(context.Background(), GrabsListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Grabs) != 1 || resp.Total != 1 {
		t.Fatal("unexpected grabs response")
	}
}

func TestWebAPI_GetPost(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	post, err := w.GetPost(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if post.PostID != "p1" || post.AuthorID != "u1" {
		t.Fatal("unexpected post")
	}
}

func TestWebAPI_GetPosts(t *testing.T) {
	ts := newWebAPITestServer(t)
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	resp, err := w.GetPosts(context.Background(), PostsListParams{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Posts) != 1 || resp.Total != 1 {
		t.Fatal("unexpected posts response")
	}
}

func TestWebAPI_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	_, err := w.GetRoom(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestWebAPI_QueryParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "5" {
			t.Fatal("missing limit param")
		}
		if r.URL.Query().Get("room_name") != "test" {
			t.Fatal("missing room_name param")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"rooms": []any{},
			"total": 0,
		})
	}))
	defer ts.Close()

	w := NewWebAPI()
	w.baseURL = ts.URL

	_, err := w.GetRooms(context.Background(), RoomsListParams{Limit: 5, RoomName: "test"})
	if err != nil {
		t.Fatal(err)
	}
}
