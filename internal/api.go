package api

import (
	"encoding/json"
	"fmt"
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
	client := &http.Client{}
	result, err := client.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product?order_by=date")
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
	client := &http.Client{}
	result, err := client.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product_cat")
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

func (c *Client) GetProduct(id int) (Product, error) {
	var product Product
	client := &http.Client{}
	result, err := client.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product/" + fmt.Sprint(id))
	if err != nil {
		return Product{}, err
	}
	defer result.Body.Close()

	err = json.NewDecoder(result.Body).Decode(&product)
	if err != nil {
		return Product{}, err
	}

	return product, nil
}

func (c *Client) GetProductsByCategory(id int) (Products, error) {
	var products Products
	client := &http.Client{}
	result, err := client.Get("https://thehoptimist.co.uk/wp-json/wp/v2/product?product_cat=" + fmt.Sprint(id))
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	err = json.NewDecoder(result.Body).Decode(&products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
