package models

type Person struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    MainImage []byte `json:"main_image"`
    Images    []Image `json:"images"`
}

type Image struct {
    ID       int    `json:"id"`
    PersonID int    `json:"person_id"`
    Image    []byte `json:"image"`
}