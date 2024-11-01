package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "grpc-product/proto/product"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

const dbPath = "product.db"

type server struct {
	pb.UnimplementedProductServiceServer
	db *sql.DB
}

func (s *server) CreateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	result, err := s.db.Exec("INSERT INTO products (name, price) VALUES (?, ?)", req.Name, req.Price)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &pb.Product{Id: id, Name: req.Name, Price: req.Price}, nil
}

func (s *server) ReadProduct(ctx context.Context, req *pb.ReadRequest) (*pb.Product, error) {
	var product pb.Product
	err := s.db.QueryRow("SELECT id, name, price FROM products WHERE id=?", req.Id).
		Scan(&product.Id, &product.Name, &product.Price)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *server) UpdateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	_, err := s.db.Exec("UPDATE products SET name=?, price=? WHERE id=?", req.Name, req.Price, req.Id)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *server) DeleteProduct(ctx context.Context, req *pb.ReadRequest) (*pb.DeleteResponse, error) {
	result, err := s.db.Exec("DELETE FROM products WHERE id=?", req.Id)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	return &pb.DeleteResponse{Success: rowsAffected > 0}, nil
}

func (s *server) ListProducts(ctx context.Context, req *pb.Empty) (*pb.ProductList, error) {
	rows, err := s.db.Query("SELECT id, name, price FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*pb.Product
	for rows.Next() {
		var product pb.Product
		if err := rows.Scan(&product.Id, &product.Name, &product.Price); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return &pb.ProductList{Products: products}, nil
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        price REAL
    )`); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, &server{db: db})

	log.Println("Server is running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
