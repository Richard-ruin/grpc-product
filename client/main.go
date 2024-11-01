package main

import (
	"context"
	"log"
	"time"

	pb "grpc-product/proto/product"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const address = "localhost:50051"

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewProductServiceClient(conn)

	// Create a new product
	product := createProduct(client, "Laptop", 1500.00)
	log.Printf("Created Product: %v\n", product)

	// Read the created product
	product = readProduct(client, product.Id)
	log.Printf("Read Product: %v\n", product)

	// Update the product
	product = updateProduct(client, product.Id, "Gaming Laptop", 1800.00)
	log.Printf("Updated Product: %v\n", product)

	// List all products
	products := listProducts(client)
	log.Printf("List of Products: %v\n", products)

	// Delete the product
	deleteProduct(client, product.Id)
	log.Printf("Deleted Product with ID: %v\n", product.Id)
}

func createProduct(client pb.ProductServiceClient, name string, price float64) *pb.Product {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	product := &pb.Product{Name: name, Price: price}
	response, err := client.CreateProduct(ctx, product)
	if err != nil {
		log.Fatalf("Error creating product: %v", err)
	}
	return response
}

func readProduct(client pb.ProductServiceClient, id int64) *pb.Product {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.ReadRequest{Id: id}
	response, err := client.ReadProduct(ctx, request)
	if err != nil {
		log.Fatalf("Error reading product: %v", err)
	}
	return response
}

func updateProduct(client pb.ProductServiceClient, id int64, name string, price float64) *pb.Product {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	product := &pb.Product{Id: id, Name: name, Price: price}
	response, err := client.UpdateProduct(ctx, product)
	if err != nil {
		log.Fatalf("Error updating product: %v", err)
	}
	return response
}

func deleteProduct(client pb.ProductServiceClient, id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.ReadRequest{Id: id}
	response, err := client.DeleteProduct(ctx, request)
	if err != nil {
		log.Fatalf("Error deleting product: %v", err)
	}
	log.Printf("Delete Response: %v", response.Success)
}

func listProducts(client pb.ProductServiceClient) []*pb.Product {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.Empty{}
	response, err := client.ListProducts(ctx, request)
	if err != nil {
		log.Fatalf("Error listing products: %v", err)
	}
	return response.Products
}
