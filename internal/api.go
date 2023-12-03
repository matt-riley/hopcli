package api

import (
	"encoding/json"
	"net/http"
)

type Product struct {
	ID          int    `json:"id"`
	Link        string `json:"link"`
	Title       string `json:"title:rendered"`
	Description string `json:"excerpt:rendered"`
}

type Products []Product

type ProductCategory struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Link  string `json:"link"`
}

type ProductCategories []ProductCategory

type Client struct{}

func (c *Client) GetLatest() (Products, error) {
	var latestProducts Products
	result, err := http.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product?order_by=date")
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	err = json.NewDecoder(result.Body).Decode(&latestProducts)
	if err != nil {
		return nil, err
	}
	return latestProducts, nil
}

func (c *Client) GetCategories() (ProductCategories, error) {
	var productCategories ProductCategories
	result, err := http.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product_cat")
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	err = json.NewDecoder(result.Body).Decode(&productCategories)
	if err != nil {
		return nil, err
	}
	return productCategories, nil
}
