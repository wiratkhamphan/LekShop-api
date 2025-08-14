package models

type PopularItem struct {
	ID       int    `json:"id"`
	SliderID int    `json:"sliderid"`
	Image    string `json:"image"`
	Alt      string `json:"alt"`
}
