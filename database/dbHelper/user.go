package dbHelper

import (
	"audio_phile/database"
	"audio_phile/model"
	"audio_phile/utils"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

func CreateUser(db sqlx.Ext, name, email, password string) (string, error) {
	SQL := `INSERT INTO users(name, email, password) VALUES ($1, TRIM(LOWER($2)), $3) RETURNING id`
	var userID string
	if err := db.QueryRowx(SQL, name, email, password).Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func IsUserExist(email string) (bool, error) {
	SQL := `SELECT id FROM users WHERE email = TRIM(LOWER($1)) AND archived_at IS NULL`
	var id string
	err := database.Audiophile.Get(&id, SQL, email)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, nil
}

func CreateUserRole(db sqlx.Ext, userID string, role model.Role) error {
	SQL := `INSERT INTO user_roles(user_id, role) VALUES ($1, $2)`
	_, err := db.Exec(SQL, userID, role)
	return err
}

//func GetAllUser() ([]model.User, error) {
//	SQL := `SELECT name,
//       			   email
//FROM
//       			            users WHERE archived_at is null`
//	list := make([]model.User, 0)
//	err := database.Audiophile.Select(&list, SQL)
//	return list, err
//}

func GetAllUser(role model.Role) ([]model.User, error) {
	SQL := `SELECT name, 
       			   email
FROM users  RIGHT JOIN user_roles on users.id = user_roles.user_id WHERE user_roles.role = $1 AND users.archived_at IS NULL `
	list := make([]model.User, 0)
	err := database.Audiophile.Select(&list, SQL, role)
	return list, err
}

func GetUserIDByEmailAndPassword(email, password string) (string, error) {
	SQL := `SELECT
				u.id,
       			u.password
       		FROM
				users u
			WHERE
				u.archived_at IS NULL
				AND u.email = TRIM(LOWER($1))`
	var userID string
	var passwordHash string
	err := database.Audiophile.QueryRowx(SQL, email).Scan(&userID, &passwordHash)
	if err != nil {
		return "", err
	}
	// compare password
	if passwordErr := utils.CheckPassword(password, passwordHash); passwordErr != nil {
		return "", passwordErr
	}
	return userID, nil
}

func GetUserRoles(userID string) (model.Role, error) {
	SQL := `SELECT role FROM user_roles WHERE user_id = $1 AND archived_at IS NULL`
	var role model.Role
	err := database.Audiophile.Get(&role, SQL, userID)
	return role, err
}

func CreateProduct(
	db sqlx.Ext,
	name, description string,
	isAvailable bool,
	price, quantity int,
	category model.Category) (string, error) {
	SQL := `INSERT INTO products(name, price, description, is_available, quantity, category) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`
	var productId string
	err := db.QueryRowx(SQL, name, price, description, isAvailable, quantity, category).Scan(&productId)
	return productId, err
}

func IsProductExist(name string) (bool, error) {
	SQL := `SELECT id FROM products WHERE name = $1 AND archived_at IS NULL`
	var id string
	err := database.Audiophile.Get(&id, SQL, name)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return true, nil
}

func GetAllProduct() ([]model.Products, error) {
	SQL := `SELECT id, 
       			   name, 
       			   price, 
       			   description, 
       			   is_available,
       			   quantity,
       			   category
FROM 
       			            products WHERE archived_at is null `

	list := make([]model.Products, 0)
	err := database.Audiophile.Select(&list, SQL)
	return list, err
}

func GetProductById(productId string) (model.Products, error) {
	SQL := `SELECT id,name,price, description,is_available, quantity, category FROM products WHERE id = $1 AND archived_at is null`
	var productModel model.Products
	err := database.Audiophile.Get(&productModel, SQL, productId)
	return productModel, err
}

func GetUserByUserId(userId string) (model.User, error) {
	SQL := `SELECT name, email FROM users WHERE id = $1 AND archived_at is null`
	var userModel model.User
	err := database.Audiophile.Get(&userModel, SQL, userId)
	return userModel, err
}

func GetAddress(db *sqlx.DB, userId string) ([]model.AddressModel, error) {
	SQL := `SELECT id, address, address_type, lat, long FROM user_addresses WHERE user_id = $1 AND archived_at is null `
	list := make([]model.AddressModel, 0)
	err := db.Select(&list, SQL, userId)
	return list, err
}

func CreateAddresses(db sqlx.Ext, userId, address string, addressType model.Address, lat, long float64) error {
	SQL := `INSERT INTO user_addresses(user_id, address, address_type, lat, long) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(SQL, userId, address, addressType, lat, long)
	return err
}

func DeleteUser(db sqlx.Ext, userId string) error {
	SQL := `UPDATE users SET archived_at = Now() WHERE id = $1`
	_, err := db.Exec(SQL, userId)
	return err
}

func CreateCart(db sqlx.Ext, userId string, status model.Status) (string, error) {
	SQL := `INSERT INTO carts(user_id, status) VALUES ($1, $2) RETURNING id`
	var cartId string
	err := db.QueryRowx(SQL, userId, status).Scan(&cartId)
	return cartId, err
}

func CreateProductInCart(db sqlx.Ext, cartId, productId string, quantity int) error {
	SQL := `INSERT INTO cart_products(cart_id, product_id, quantity) VALUES ($1, $2, $3)`
	_, err := db.Exec(SQL, cartId, productId, quantity)
	return err

}

func IsCartExist(userId string, cartStatus model.Status) (string, bool, error) {
	SQL := `SELECT id FROM carts WHERE user_id = $1 AND status = $2`
	var id string
	err := database.Audiophile.Get(&id, SQL, userId, cartStatus)
	if err != nil && err != sql.ErrNoRows {
		return "", false, err
	}
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	return id, true, nil
}

//func GetCartWithProduct(db *sqlx.DB, cartId string) ([]model.CartProduct, error) {
//	SQL := `SELECT product_id FROM cart_products WHERE cart_id = $1`
//	list := make([]model.CartProduct, 0)
//	err := db.Select(&list, SQL, cartId)
//	return list, err
//}

func GetCartWithProduct(db *sqlx.DB, userId string) ([]model.CartProduct, error) {
	SQL := `SELECT p.id,
       p.name,
       p.price,
       p.description,
       cp.quantity
FROM products p INNER JOIN cart_products cp ON p.id = cp.product_id INNER JOIN carts c on cp.cart_id = c.id WHERE user_id = $1 AND cp.archived_at IS NULL `
	list := make([]model.CartProduct, 0)
	err := db.Select(&list, SQL, userId)
	return list, err
}

func GetCartProductQuantity(cartId string, productId string) (model.QuantityOfProductInCart, error) {
	SQL := `SELECT cart_id, product_id, quantity FROM cart_products WHERE cart_id = $1 AND product_id = $2`
	var cartDetail model.QuantityOfProductInCart
	err := database.Audiophile.Get(&cartDetail, SQL, cartId, productId)
	return cartDetail, err

}

func AddProductQuantityInCart(cartId, productId string) error {
	SQL := `UPDATE cart_products SET quantity = cart_products.quantity + 1 WHERE cart_id = $1 AND product_id = $2`
	_, err := database.Audiophile.Exec(SQL, cartId, productId)
	return err
}

func RemoveProductQuantityInCart(cartId, productId string) error {
	SQL := `UPDATE cart_products SET quantity = cart_products.quantity - 1 WHERE cart_id = $1 AND product_id = $2`
	_, err := database.Audiophile.Exec(SQL, cartId, productId)
	return err
}

func UpdateProductFromCart(db sqlx.Ext, cartId, productId string) error {
	SQL := `UPDATE cart_products SET archived_at = Now() WHERE cart_id = $1 AND product_id = $2`
	_, err := db.Exec(SQL, cartId, productId)
	return err
}

func CreateOrder(db sqlx.Ext, cartProductId string, orderStatus model.OrderStatus, addressId string) error {
	SQL := `INSERT INTO orders(cart_id, order_status, address_id) VALUES ($1, $2, $3)`
	_, err := db.Exec(SQL, cartProductId, orderStatus, addressId)
	return err
}
func GetCartProductByID(cartId string) ([]model.ProductMinimalDetails, error) {
	SQL := `SELECT product_id, quantity FROM cart_products WHERE cart_id = $1 AND archived_at IS NULL`
	list := make([]model.ProductMinimalDetails, 0)
	err := database.Audiophile.Select(&list, SQL, cartId)
	return list, err
}

func UpdateProductQuantity(db sqlx.Ext, productId string, quantity int) error {
	SQL := `UPDATE products SET quantity = products.quantity - $2 WHERE id = $1`
	_, err := db.Exec(SQL, productId, quantity)
	return err
}

func IsCartIsActive(cartId string) (model.Status, error) {
	SQL := `SELECT status FROM carts WHERE id = $1`
	var status model.Status
	err := database.Audiophile.Get(&status, SQL, cartId)
	return status, err
}

func UpdateCartToInactive(db sqlx.Ext, cartId string, status model.Status) error {
	SQL := `UPDATE carts SET status = $1 WHERE id = $2`
	_, err := db.Exec(SQL, status, cartId)
	return err
}

func CreateOrderStatus(db sqlx.Ext, orderId string, status model.OrderStatus) error {
	SQL := `UPDATE orders SET order_status = $1 WHERE id = $2`
	_, err := db.Exec(SQL, status, orderId)
	return err
}

func UploadImageFirebase(bucket, imagePath string) (string, error) {
	SQL := `INSERT INTO attachments(image_path, bucket_name) VALUES ($1, $2) RETURNING id`
	var imageID string
	if err := database.Audiophile.QueryRowx(SQL, imagePath, bucket).Scan(&imageID); err != nil {
		return "", err
	}
	return imageID, nil
}
