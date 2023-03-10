package handler

import (
	"audio_phile/database"
	"audio_phile/database/dbHelper"
	"audio_phile/middleware"
	"audio_phile/model"
	"audio_phile/utils"
	cloud "cloud.google.com/go/storage"
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

func UploadImages(ctx *gin.Context) {
	client := model.FirebaseClient

	var file multipart.File
	var fileHeader *multipart.FileHeader
	var err error

	file, fileHeader, err = ctx.Request.FormFile("image")
	err = ctx.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		logrus.Errorf("UploadImages: error in parsing multipart form err = %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "error in parsing multipart form"})
		return
	}

	defer file.Close()
	imagePath := fileHeader.Filename + strconv.Itoa(int(time.Now().Unix()))
	bucket := "audiophile-c47c3.appspot.com"
	bucketStorage := client.Storage.Bucket(bucket).Object(imagePath).NewWriter(client.Ctx)

	_, err = io.Copy(bucketStorage, file)
	if err != nil {
		logrus.Errorf("UploadImages: error in file copying err: %v", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "message": "error in file copying err"})
		return
	}

	productID := ctx.Param("productID")

	var imageID string
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		imageID, err = dbHelper.UploadImageFirebase(tx, bucket, imagePath)
		if err != nil {
			logrus.Errorf("UploadImages: error in uploading image to firebase err = %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in uploading image to firebase"})
			return err
		}

		err = dbHelper.CreateProductAttachments(tx, imageID, productID)
		if err != nil {
			logrus.Errorf("UploadImages: error in uploading image to firebase err = %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in uploading image to firebase"})
			return err
		}
		return nil
	})
	if txErr != nil {
		logrus.Errorf("Transaction: error in transaction err = %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error(), "message": "Failed to create user"})
		return
	}

	if err := bucketStorage.Close(); err != nil {
		logrus.Errorf("UploadImages: error in closing firebase bucket err = %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error(), "message": "error in closing firebase bucket"})
		return
	}

	ctx.JSON(http.StatusOK, struct {
		ImageID string
	}{imageID})
}

func CreateUser(ctx *gin.Context) {
	var body model.UserRequestBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Failed to parse request body"})
		return
	}

	parseBody := model.UserRequestBody{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Invalid inputs"})
		return
	}

	if len(body.Password) < 6 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "password must be 6 chars long"})
		return
	}

	exist, existErr := dbHelper.IsUserExist(body.Email)
	if exist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": existErr.Error(), "message": "User already exist"})
		return
	}

	if existErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": existErr.Error(), "message": "Failed to check existence"})
		return
	}

	hashPassword, hasErr := utils.HashPassword(body.Password)

	if hasErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": hasErr.Error(), "message": "Failed to secure password"})
		return
	}

	var userID string
	var err error

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userID, err = dbHelper.CreateUser(tx, body.Name, body.Email, hashPassword)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": hasErr.Error(), "message": "Failed to create user"})
			return err
		}

		roleErr := dbHelper.CreateUserRole(tx, userID, model.RoleUser)
		if roleErr != nil {
			return roleErr
		}
		return nil
	})
	// error message correct karo
	if txErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error(), "message": "Transaction error"})
		return
	}

	ctx.JSON(http.StatusCreated, model.UserResponseBody{
		UserId: userID,
		Name:   body.Name,
		Email:  body.Email,
	})

}

func Login(ctx *gin.Context) {
	var body model.LoginRequestBody

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Failed to parse request body",
		})
		return
	}

	parseBody := model.LoginRequestBody{
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Input field is invalid",
		})
		return
	}

	userId, err := dbHelper.GetUserIDByEmailAndPassword(body.Email, body.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "user does not exist",
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Incorrect credentials",
		})
		return
	}

	role, err := dbHelper.GetUserRoles(userId)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Error in getting user role",
		})
		return
	}

	token, err := middleware.GenerateJWT(userId, role)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Error in generating jwt token",
		})
		return
	}
	ctx.JSON(http.StatusCreated, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func CreateProduct(ctx *gin.Context) {
	var body model.ProductsRequest

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Failed to parse request body"})
		return
	}

	parseBody := model.ProductsRequest{
		Name:        body.Name,
		Price:       body.Price,
		Description: body.Description,
		IsAvailable: body.IsAvailable,
		Quantity:    body.Quantity,
		Category:    body.Category,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Input field is invalid"})
		return
	}

	exist, existErr := dbHelper.IsProductExist(body.Name)
	if exist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": existErr.Error(), "message": "Product already exist"})
		return
	}

	if existErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": existErr.Error(), "message": "Failed to product existence"})
		return
	}

	productId, err := dbHelper.CreateProduct(database.Audiophile, body.Name, body.Description, body.IsAvailable, body.Price, body.Quantity, body.Category)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": existErr.Error(), "message": "Failed to create product"})
		return
	}

	ctx.JSON(http.StatusCreated, model.ProductsResponse{
		ProductId:   productId,
		Name:        body.Name,
		Price:       body.Price,
		Description: body.Description,
		IsAvailable: body.IsAvailable,
		Quantity:    body.Quantity,
		Category:    body.Category,
	})
}

func GetAllProduct(ctx *gin.Context) {
	list, err := dbHelper.GetAllProductWithImage()
	logrus.Println(list)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Failed to fetch all product"})
		return
	}
	client := model.FirebaseClient
	for _, product := range list {
		signedUrl := &cloud.SignedURLOptions{
			Scheme:  cloud.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(15 * time.Minute),
		}
		url, err := client.Storage.Bucket(product.BucketName).SignedURL(product.Path, signedUrl)
		if err != nil {
			logrus.Errorf("GetAllProducts: error in generating image url err: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "error in generating image url"})
			return
		}

		ctx.JSON(http.StatusCreated, struct {
			Id          string
			Name        string
			Price       int
			Description string
			IsAvailable bool
			Quantity    int
			Category    model.Category
			ImageUrl    string
		}{
			Id:          product.ProductId,
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			IsAvailable: product.IsAvailable,
			Quantity:    product.Quantity,
			Category:    product.Category,
			ImageUrl:    url,
		})
	}
}

func GetProductById(ctx *gin.Context) {
	productId := ctx.Param("id")
	var product model.Products
	var productDetails model.ProductDetails
	var err error

	product, err = dbHelper.GetProductById(productId)
	if err != nil {
		logrus.Errorf("Get Product Detail: error in getting product details: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Error in fetching product details"})
		return
	}

	var imageDetail []model.Images
	imgSlice := make([]string, 0)

	imageDetail, err = dbHelper.GetImageByProductID(productId)

	if err != nil {
		logrus.Errorf("Product Image: error in generating image url err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Error in getting image"})
		return
	}

	client := model.FirebaseClient

	for _, product := range imageDetail {
		signedUrl := &cloud.SignedURLOptions{
			Scheme:  cloud.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(15 * time.Minute),
		}
		url, err := client.Storage.Bucket(product.BucketName).SignedURL(product.ImagePath, signedUrl)
		if err != nil {
			logrus.Errorf("GetAllProducts: error in generating image url err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Error in generating image url err"})
			return
		}
		imgSlice = append(imgSlice, url)
	}

	productDetails.ProductId = product.ProductId
	productDetails.Name = product.Name
	productDetails.Description = product.Description
	productDetails.Category = product.Category
	productDetails.Quantity = product.Quantity
	productDetails.IsAvailable = product.IsAvailable
	productDetails.Price = product.Price
	productDetails.ImageUrl = imgSlice

	ctx.JSON(http.StatusCreated, productDetails)
}

func CreatedAddress(ctx *gin.Context) {
	// parse the request data
	var addresses []model.AddressRequest
	if err := ctx.ShouldBindJSON(&addresses); err != nil {
		logrus.Errorf("Create Address: error in creating address err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to parse request body"})
		return
	}
	userId := getUserId(ctx)

	for _, address := range addresses {
		// save the address to the database
		err := dbHelper.CreateAddresses(database.Audiophile, userId, address.Address, address.AddressType, address.Lat, address.Long)
		if err != nil {
			logrus.Errorf("Create Role: error in creating role err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create role"})
			return
		}
	}

	ctx.JSON(http.StatusCreated, struct {
		Message string
	}{"Address created successfully!!"})
}

func GetUserByUserId(ctx *gin.Context) {
	userId := ctx.Param("id")
	var userData model.User
	var userAddress []model.AddressModel
	var userDetails model.UserWithAddress
	var err error
	userData, err = dbHelper.GetUserByUserId(userId)
	if err != nil {
		return
	}
	userAddress, err = dbHelper.GetAddress(database.Audiophile, userId)
	if err != nil {
		return
	}
	userDetails.Id = userId
	userDetails.Name = userData.Name
	userDetails.Email = userData.Email
	userDetails.Address = userAddress
	ctx.JSON(http.StatusCreated, userDetails)
}

//func getUserId(ctx *gin.Context) string {
//	user := ctx.Request.Context().Value(middleware.UserContext).(map[string]interface{})
//	fmt.Println(user)
//	var userId string
//	userId = user["id"].(string)
//	fmt.Println(userId)
//	return userId
//}

func getUserId(c *gin.Context) string {
	user, exists := c.Get(string(middleware.UserContext))
	if !exists {
		return ""
	}
	userData, ok := user.(map[string]interface{})
	if !ok {
		return ""
	}
	userId, ok := userData["id"].(string)
	if !ok {
		return ""
	}
	return userId
}

func GetAllUser(ctx *gin.Context) {
	list, err := dbHelper.GetAllUser(model.RoleUser)
	logrus.Println(list)
	if err != nil {
		logrus.Errorf("Get All user: error in getting all user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Error in getting all user"})
		return
	}
	ctx.AsciiJSON(http.StatusOK, list)
}

func DeleteUserByUserId(ctx *gin.Context) {
	userId := ctx.Param("id")
	err := dbHelper.DeleteUser(database.Audiophile, userId)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusOK, struct {
		Message string
	}{"User deleted successfully!"})
}

func CreateProductToCart(ctx *gin.Context) {
	userId := getUserId(ctx)
	productId := ctx.Param("id")
	quantityStr := ctx.Param("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		logrus.Errorf("Create Product to cart: error in creating product in cart: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to covert from string to int"})
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		logrus.Errorf("GetProductById: error in getting product by id err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Product not found!"})
		return
	}

	if quantity > productDetail.Quantity {
		logrus.Errorf("Product Quantity : error in product quantity err: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Requested quantity not available"})
		return
	}

	existingCartId, exist, existErr := dbHelper.IsCartExist(userId, model.CartStatusActive)

	if existErr != nil {
		logrus.Errorf("IsCartExist : error in cart exist err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to check cart existence"})
		return
	}

	if exist {
		err := dbHelper.CreateProductInCart(database.Audiophile, existingCartId, productId, quantity)
		if err != nil {
			logrus.Errorf("CreateProductInCart : error in creating product in cart err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to add product"})
			return
		}
	} else {
		txErr := database.Tx(func(tx *sqlx.Tx) error {
			cartId, err := dbHelper.CreateCart(tx, userId, model.CartStatusActive)
			if err != nil {
				logrus.Errorf("CreateCart : error in creating cart err: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create cart"})
				return err
			}

			err = dbHelper.CreateProductInCart(tx, cartId, productId, quantity)
			if err != nil {
				logrus.Errorf("Add Product : error in adding product to cart err: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to add product"})
				return err
			}
			return nil
		})
		// error message correct karo
		if txErr != nil {
			logrus.Errorf("Transaction : error in transaction err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error(), "message": "Transaction error"})
			return
		}
	}

	ctx.JSON(http.StatusCreated, struct {
		Message string
	}{"Product added to cart!"})
}

func GetCartWithProductById(ctx *gin.Context) {
	userId := getUserId(ctx)
	list, err := dbHelper.GetCartWithProduct(database.Audiophile, userId)
	logrus.Println(list)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, list)
}

func AddProductQuantityInCart(ctx *gin.Context) {
	cartProductId := ctx.Param("cartId")
	productId := ctx.Param("productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		logrus.Errorf("Transaction : error in transaction err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Cart detail not found!"})
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		logrus.Errorf("Product Detial : error in gettng product detail err: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Product not found!"})
		return
	}

	if cartDetail.Quantity >= productDetail.Quantity {
		logrus.Errorf("Product Quantity : error in quantity err: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Requested quantity not available"})
		return
	}

	err = dbHelper.AddProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})
}

func RemoveProductQuantityInCart(ctx *gin.Context) {
	cartProductId := ctx.Param("cartId")
	productId := ctx.Param("productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		logrus.Errorf("CartDetails : error in cart details err: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Cart detail not found!"})
		return
	}

	if cartDetail.Quantity == 0 {
		logrus.Errorf("Product Quantity : error in product quantity err: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Quantity is already 0"})
		return
	}

	err = dbHelper.RemoveProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})

}

func DeleteProductFromCart(ctx *gin.Context) {
	cartProductId := ctx.Param("cartId")
	productId := ctx.Param("productId")

	err := dbHelper.UpdateProductFromCart(database.Audiophile, cartProductId, productId)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Message string
	}{"Product deleted successfully"})
}

func CreateOrder(ctx *gin.Context) {
	cartId := ctx.Param("cartId")
	addressId := ctx.Param("addressId")

	active, activeErr := dbHelper.IsCartIsActive(cartId)

	if active == model.CartStatusInActive {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": activeErr.Error(), "message": "Cart is not active"})
		return
	}

	if activeErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": activeErr.Error(), "message": "Failed to check cart status"})
		return
	}

	productDetail, err := dbHelper.GetCartProductByID(cartId)
	if err != nil {
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := dbHelper.CreateOrder(tx, cartId, model.OrderStatusOrdered, addressId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create order"})
			return err
		}

		for _, v := range productDetail {
			err = dbHelper.UpdateProductQuantity(tx, v.ProductId, v.Quantity)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to update the quantity"})
				return err
			}

			err := dbHelper.UpdateProductFromCart(tx, cartId, v.ProductId)

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to delete product from cart"})
				return err
			}
		}

		err = dbHelper.UpdateCartToInactive(tx, cartId, model.CartStatusInActive)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to update status of cart"})
			return err
		}

		return nil
	})
	// error message correct karo
	if txErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error(), "message": "transaction error"})
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Message string
	}{Message: "Order placed successfully"})
}

func CreateOrderStatus(ctx *gin.Context) {
	orderId := ctx.Param("orderId")
	status := ctx.Param("orderStatus")
	var orderStatus model.OrderStatus
	if status == "shipping" {
		orderStatus = model.OrderStatusShipping
	} else if status == "delivered" {
		orderStatus = model.OrderStatusDelivered
	}
	err := dbHelper.CreateOrderStatus(database.Audiophile, orderId, orderStatus)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create order status"})
		return
	}

	ctx.JSON(http.StatusOK, struct {
		Message string
	}{"order status changed"})
}

func GetUserAddress(ctx *gin.Context) {
	userId := getUserId(ctx)
	userAddress, err := dbHelper.GetAddress(database.Audiophile, userId)
	logrus.Println(userAddress)
	if err != nil {
		return
	}
	ctx.JSON(http.StatusCreated, userAddress)
}

func GetAllImageByProductId(ctx *gin.Context) {
	productID := ctx.Param("productID")
	imageDetails, err := dbHelper.GetImageByProductID(productID)

	if err != nil {
		logrus.Errorf("FetchImages: error in getting image err = %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in getting image"})
		return
	}

	client := model.FirebaseClient

	for _, product := range imageDetails {
		signedUrl := &cloud.SignedURLOptions{
			Scheme:  cloud.SigningSchemeV4,
			Method:  "GET",
			Expires: time.Now().Add(15 * time.Minute),
		}
		url, err := client.Storage.Bucket(product.BucketName).SignedURL(product.ImagePath, signedUrl)
		if err != nil {
			logrus.Errorf("GetAllProducts: error in generating image url err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "error in generating image url"})
			return
		}

		ctx.JSON(http.StatusOK, struct {
			Id       string
			ImageUrl string
		}{
			Id:       product.ImageID,
			ImageUrl: url,
		})
	}
}

//func GetCartProductIds(w http.ResponseWriter, r *http.Request) {
//	cartProductId := chi.URLParam(r, "cartId")
//	body, err := dbHelper.GetCartProductIdByID(cartProductId)
//	if err != nil {
//		return
//	}
//	utils.RespondJSON(w, http.StatusOK, body)
//}

//func UpdateProductQuantity(w http.ResponseWriter, r *http.Request) {
//	productId := chi.URLParam(r, "productId")
//	quantityStr := chi.URLParam(r, "quantity")
//	quantity, err := strconv.Atoi(quantityStr)
//
//	if err != nil {
//		utils.RespondError(w, http.StatusInternalServerError, err, "error in fetching quantity")
//		return
//	}
//
//	err = dbHelper.UpdateProductQuantity(productId, quantity)
//
//	if err != nil {
//		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update the quantity")
//		return
//	}
//
//	utils.RespondJSON(w, http.StatusOK, struct {
//		Message string
//	}{"Product quantity updated successfully"})
//}
