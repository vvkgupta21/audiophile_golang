package model

type Role string
type Category string
type Address string
type Status string
type OrderStatus string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

const (
	CategoryHeadphones Category = "headphones"
	CategorySpeakers   Category = "speakers"
	CategoryEarphones  Category = "earphones"
)

const (
	AddressHome   Address = "home"
	AddressOffice Address = "office"
	AddressOther  Address = "other"
)
const (
	CartStatusActive   Status = "active"
	CartStatusInActive Status = "inactive"
)

const (
	OrderStatusOrdered   OrderStatus = "ordered"
	OrderStatusShipping  OrderStatus = "shipping"
	OrderStatusDelivered OrderStatus = "delivered"
)

type UserRequestBody struct {
	Name     string `json:"name" db:"name" validate:"required,min=3,max=15"`
	Email    string `json:"email" db:"email" validate:"required,email"`
	Password string `json:"password" db:"password" validate:"omitempty,len=6,numeric"`
}

type UserResponseBody struct {
	UserId string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type LoginRequestBody struct {
	Email    string `json:"email" db:"email" validate:"required,email"`
	Password string `json:"password" db:"password" validate:"omitempty,len=6,numeric"`
}

type UserCredential struct {
	Id    string `json:"id"`
	Roles Role   `json:"role"`
}

type ProductsRequest struct {
	Name        string   `json:"name" db:"name" validate:"required"`
	Price       int      `json:"price" db:"price" validate:"required"`
	Description string   `json:"description" db:"description" validate:"required"`
	IsAvailable bool     `json:"is_available" db:"is_available" validate:"required"`
	Quantity    int      `json:"quantity" db:"quantity"`
	Category    Category `json:"category" db:"category" validate:"required"`
}

type ProductsResponse struct {
	ProductId   string   `json:"productId"`
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Description string   `json:"description"`
	IsAvailable bool     `json:"is_available"`
	Quantity    int      `json:"quantity"`
	Category    Category `json:"category"`
}

type Products struct {
	ProductId   string   `json:"productId" db:"id"`
	Name        string   `json:"name" db:"name"`
	Price       int      `json:"price" db:"price"`
	Description string   `json:"description" db:"description"`
	IsAvailable bool     `json:"isAvailable" db:"is_available"`
	Quantity    int      `json:"quantity" db:"quantity"`
	Category    Category `json:"category" db:"category"`
	BucketName  string   `json:"bucket_name" db:"bucket_name"`
	Path        string   `json:"image_path" db:"image_path"`
}

type Images struct {
	ImageID   string `json:"imageID" db:"id"`
	ImagePath string `json:"imagePath" db:"image_path"`
}

type User struct {
	Name  string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
}

type AddressRequest struct {
	UserID      string  `json:"userID" db:"user_id"`
	Address     string  `json:"address" db:"address"`
	AddressType Address `json:"address_type" db:"address_type"`
	Lat         float64 `json:"lat" db:"lat"`
	Long        float64 `json:"long" db:"long"`
}

type AddressModel struct {
	Id          string  `json:"id" db:"id"`
	Address     string  `json:"address" db:"address"`
	AddressType Address `json:"address_type" db:"address_type"`
	Lat         float64 `json:"lat" db:"lat"`
	Long        float64 `json:"long" db:"long"`
}

type UserWithAddress struct {
	Id      string         `json:"id"`
	Name    string         `json:"name"`
	Email   string         `json:"email"`
	Address []AddressModel `json:"address"`
}

type CartModel struct {
	Id     string `json:"id" db:"id"`
	UserId string `json:"userId" db:"user_id"`
}

type CartProduct struct {
	ProductId   string `json:"productId" db:"id"`
	Name        string `json:"name" dbv:"name"`
	Price       int    `json:"price" db:"price"`
	Description string `json:"description" db:"description"`
	Quantity    int    `json:"quantity" db:"quantity"`
}

type QuantityOfProductInCart struct {
	CartId    string `json:"cartId" db:"cart_id"`
	ProductId string `json:"productId" db:"product_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
}

type ProductMinimalDetails struct {
	ProductId string `json:"product_id" db:"product_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
}

type PlacedOrderStatus struct {
	Status OrderStatus `json:"order_status" db:"order_status"`
}
