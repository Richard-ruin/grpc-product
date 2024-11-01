package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "grpc-product/proto/product"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var tmpl *template.Template
var client pb.ProductServiceClient

func init() {
	tmpl = template.Must(template.ParseGlob("web/templates/*.html"))
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client = pb.NewProductServiceClient(conn)

	http.HandleFunc("/", listProducts)
	http.HandleFunc("/create", createProductForm)
	http.HandleFunc("/edit", editProductForm)
	http.HandleFunc("/save", saveProduct)
	http.HandleFunc("/delete", deleteProduct)

	log.Println("Starting web server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listProducts(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.ListProducts(ctx, &pb.Empty{})
	if err != nil {
		http.Error(w, "Unable to fetch products", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "index.html", res.Products)
}

func createProductForm(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "form.html", nil)
}

func editProductForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.ReadProduct(ctx, &pb.ReadRequest{Id: id})
	if err != nil {
		http.Error(w, "Unable to fetch product", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "form.html", res)
}

func saveProduct(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	name := r.FormValue("name")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	product := &pb.Product{
		Id:    id,
		Name:  name,
		Price: price,
	}

	if id == 0 {
		// Create new product
		_, err := client.CreateProduct(ctx, product)
		if err != nil {
			http.Error(w, "Unable to create product", http.StatusInternalServerError)
			return
		}
	} else {
		// Update existing product
		_, err := client.UpdateProduct(ctx, product)
		if err != nil {
			http.Error(w, "Unable to update product", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := client.DeleteProduct(ctx, &pb.ReadRequest{Id: id})
	if err != nil {
		http.Error(w, "Unable to delete product", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
