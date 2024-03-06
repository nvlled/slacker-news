package hn

type ItemID int64

type Item struct {
	ID          ItemID   `json:"id"`
	Deleted     bool     `json:"deleted"`
	Type        string   `json:"type"`
	By          string   `json:"by"`
	Time        int64    `json:"time"`
	Text        string   `json:"text"`
	Dead        bool     `json:"dead"`
	Parent      ItemID   `json:"parent"`
	Poll        string   `json:"poll"`
	Kids        []ItemID `json:"kids"`
	Url         string   `json:"url"`
	Score       int      `json:"score"`
	Title       string   `json:"title"`
	Parts       []int    `json:"parts"`
	Descendants int      `json:"descendants"`


        // Added non HN fields
        Level uint
        ParentItem *Item
}
