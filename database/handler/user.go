package handler

import (
	"audio_phile/database"
	"audio_phile/database/dbHelper"
	"audio_phile/middleware"
	"audio_phile/model"
	"audio_phile/utils"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"time"
)

//func UploadImage(writer http.ResponseWriter, request *http.Request) {
//	client := model.App{}
//	var err error
//	client.Ctx = context.Background()
//	credentialsFile := option.WithCredentialsJSON([]byte(os.Getenv("Firebase_Storage_Credential")))
//	app, err := firebase.NewApp(client.Ctx, nil, credentialsFile)
//	if err != nil {
//		logrus.Error(err)
//		return
//	}
//
//	client.Client, err = app.Firestore(client.Ctx)
//	if err != nil {
//		logrus.Error(err)
//		return
//	}
//
//	client.Storage, err = cloud.NewClient(client.Ctx, credentialsFile)
//	if err != nil {
//		logrus.Error(err)
//		return
//	}
//
//	file, fileHeader, err := request.FormFile("image")
//	err = request.ParseMultipartForm(10 << 20)
//	if err != nil {
//		logrus.Error(err)
//		writer.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//
//	defer file.Close()
//	imagePath := fileHeader.Filename + strconv.Itoa(int(time.Now().Unix()))
//	bucket := "audiophile-c47c3.appspot.com"
//	bucketStorage := client.Storage.Bucket(bucket).Object(imagePath).NewWriter(client.Ctx)
//
//	_, err = io.Copy(bucketStorage, file)
//	if err != nil {
//		logrus.Error(err)
//		writer.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	if err := bucketStorage.Close(); err != nil {
//		logrus.Error(err)
//		writer.WriteHeader(http.StatusBadRequest)
//		return
//	}
//
//	signedUrl := &cloud.SignedURLOptions{
//		Scheme:  cloud.SigningSchemeV4,
//		Method:  "GET",
//		Expires: time.Now().Add(15 * time.Minute),
//	}
//	url, err := client.Storage.Bucket(bucket).SignedURL(imagePath, signedUrl)
//	if err != nil {
//		logrus.Error(err)
//		return
//	}
//	logrus.Println(url)
//	errs := json.NewEncoder(writer).Encode(url)
//	if errs != nil {
//		logrus.Error(err)
//		writer.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//}

func UploadImages(w http.ResponseWriter, r *http.Request) {
	client := model.FirebaseClient

	file, fileHeader, err := r.FormFile("image")
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		logrus.Errorf("UploadImages: error in parsing multipart form err = %v", err)
		utils.RespondError(w, http.StatusInternalServerError, err, "error in parsing multipart form")
		return
	}

	defer file.Close()
	imagePath := fileHeader.Filename + strconv.Itoa(int(time.Now().Unix()))
	bucket := "audiophile-c47c3.appspot.com"
	bucketStorage := client.Storage.Bucket(bucket).Object(imagePath).NewWriter(client.Ctx)

	_, err = io.Copy(bucketStorage, file)
	if err != nil {
		logrus.Errorf("UploadImages: error in file copying err: %v", err)
		utils.RespondError(w, http.StatusBadGateway, err, "error in file copying err")
		return
	}

	imageId, err := dbHelper.UploadImageFirebase(bucket, imagePath)
	if err != nil {
		logrus.Errorf("UploadImages: error in uploading image to firebase err = %v", err)
		utils.RespondError(w, http.StatusInternalServerError, err, "error in uploading image to firebase")
		return
	}

	if err := bucketStorage.Close(); err != nil {
		logrus.Errorf("UploadImages: error in closing firebase bucket err = %v", err)
		utils.RespondError(w, http.StatusInternalServerError, err, "error in closing firebase bucket")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, imageId)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var body model.UserRequestBody

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	parseBody := model.UserRequestBody{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	if len(body.Password) < 6 {
		utils.RespondError(w, http.StatusBadRequest, nil, "password must be 6 chars long")
		return
	}

	exist, existErr := dbHelper.IsUserExist(body.Email)
	if exist {
		utils.RespondError(w, http.StatusBadRequest, existErr, "User already exist")
		return
	}

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to check existence")
		return
	}

	hashPassword, hasErr := utils.HashPassword(body.Password)

	if hasErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, hasErr, "failed to secure password")
		return
	}

	var userID string
	var err error

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userID, err = dbHelper.CreateUser(tx, body.Name, body.Email, hashPassword)

		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create user")
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
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to create user")
		return
	}

	//code could be 201
	utils.RespondJSON(w, http.StatusCreated, model.UserResponseBody{
		UserId: userID,
		Name:   body.Name,
		Email:  body.Email,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var body model.LoginRequestBody

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}

	parseBody := model.LoginRequestBody{
		Email:    body.Email,
		Password: body.Password,
	}
	validate := validator.New()
	if err := validate.Struct(parseBody); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	userId, err := dbHelper.GetUserIDByEmailAndPassword(body.Email, body.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.RespondError(w, http.StatusBadRequest, errors.New("user does not exist"), "user does not exist")
			return
		}
		utils.RespondError(w, http.StatusBadRequest, err, "Incorrect credentials")
		return
	}

	role, err := dbHelper.GetUserRoles(userId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error in getting user role")
		return
	}

	token, err := middleware.GenerateJWT(userId, role)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "error in generating jwt token")
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var body model.ProductsRequest

	if err := utils.ParseBody(r.Body, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
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
		utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
		return
	}

	exist, existErr := dbHelper.IsProductExist(body.Name)
	if exist {
		utils.RespondError(w, http.StatusBadRequest, existErr, "Product already exist")
		return
	}

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to product existence")
		return
	}

	productId, err := dbHelper.CreateProduct(database.Audiophile, body.Name, body.Description, body.IsAvailable, body.Price, body.Quantity, body.Category)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create product")
		return
	}

	//code could be 201
	utils.RespondJSON(w, http.StatusCreated, model.ProductsResponse{
		ProductId:   productId,
		Name:        body.Name,
		Price:       body.Price,
		Description: body.Description,
		IsAvailable: body.IsAvailable,
		Quantity:    body.Quantity,
		Category:    body.Category,
	})
}

func GetAllProduct(w http.ResponseWriter, r *http.Request) {
	list, err := dbHelper.GetAllProduct()
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func GetProductById(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		return
	}
	utils.RespondJSON(w, http.StatusOK, productDetail)
}

func CreatedAddress(w http.ResponseWriter, r *http.Request) {
	// parse the request data
	var addresses []model.AddressRequest
	if err := utils.ParseBody(r.Body, &addresses); err != nil {
		utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
		return
	}
	userId := getUserId(r)
	for _, address := range addresses {
		// save the address to the database
		err := dbHelper.CreateAddresses(database.Audiophile, userId, address.Address, address.AddressType, address.Lat, address.Long)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create role")
			return
		}
	}
	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string
	}{"Address created successfully!!"})
}

func GetUserByUserId(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
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
	utils.RespondJSON(w, http.StatusOK, userDetails)
}

func getUserId(r *http.Request) string {
	user := r.Context().Value(middleware.UserContext).(map[string]interface{})
	fmt.Println(user)
	var userId string
	userId = user["id"].(string)
	fmt.Println(userId)
	return userId
}

func GetAllUser(w http.ResponseWriter, r *http.Request) {
	list, err := dbHelper.GetAllUser(model.RoleUser)
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func DeleteUserByUserId(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	err := dbHelper.DeleteUser(database.Audiophile, userId)
	if err != nil {
		return
	}
	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"User deleted successfully!"})
}

func CreateProductToCart(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(r)
	productId := chi.URLParam(r, "id")
	quantityStr := chi.URLParam(r, "quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to covert from string to int")
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Product not found!")
		return
	}

	if quantity > productDetail.Quantity {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	existingCartId, exist, existErr := dbHelper.IsCartExist(userId, model.CartStatusActive)

	if existErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existErr, "Failed to check cart existence")
		return
	}

	if exist {
		err := dbHelper.CreateProductInCart(database.Audiophile, existingCartId, productId, quantity)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to add product")
			return
		}
	} else {
		txErr := database.Tx(func(tx *sqlx.Tx) error {
			cartId, err := dbHelper.CreateCart(tx, userId, model.CartStatusActive)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create cart")
				return err
			}

			err = dbHelper.CreateProductInCart(tx, cartId, productId, quantity)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to add product")
				return err
			}
			return nil
		})
		// error message correct karo
		if txErr != nil {
			utils.RespondError(w, http.StatusInternalServerError, txErr, "transaction error")
			return
		}
	}

	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string
	}{"Product added to cart!"})
}

func GetCartWithProductById(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(r)
	list, err := dbHelper.GetCartWithProduct(database.Audiophile, userId)
	logrus.Println(list)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, list)
	if err != nil {
		return
	}
}

func AddProductQuantityInCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Cart detail not found!")
		return
	}

	productDetail, err := dbHelper.GetProductById(productId)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Product not found!")
		return
	}

	if cartDetail.Quantity >= productDetail.Quantity {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	err = dbHelper.AddProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})
}

func RemoveProductQuantityInCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	cartDetail, err := dbHelper.GetCartProductQuantity(cartProductId, productId)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Cart detail not found!")
		return
	}

	if cartDetail.Quantity <= 0 {
		utils.RespondError(w, http.StatusBadRequest, nil, "Requested quantity not available")
		return
	}

	err = dbHelper.RemoveProductQuantityInCart(cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Quantity updated successfully"})

}

func DeleteProductFromCart(w http.ResponseWriter, r *http.Request) {
	cartProductId := chi.URLParam(r, "cartId")
	productId := chi.URLParam(r, "productId")

	err := dbHelper.UpdateProductFromCart(database.Audiophile, cartProductId, productId)
	if err != nil {
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{"Product deleted successfully"})
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	cartId := chi.URLParam(r, "cartId")
	addressId := chi.URLParam(r, "addressId")

	active, activeErr := dbHelper.IsCartIsActive(cartId)

	if active == model.CartStatusInActive {
		utils.RespondError(w, http.StatusBadRequest, activeErr, "Cart is not active")
		return
	}

	if activeErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, activeErr, "Failed to check cart status")
		return
	}

	productDetail, err := dbHelper.GetCartProductByID(cartId)
	if err != nil {
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := dbHelper.CreateOrder(tx, cartId, model.OrderStatusOrdered, addressId)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create order")
			return err
		}

		for _, v := range productDetail {
			err = dbHelper.UpdateProductQuantity(tx, v.ProductId, v.Quantity)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update the quantity")
				return err
			}

			err := dbHelper.UpdateProductFromCart(tx, cartId, v.ProductId)

			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, err, "Failed to delete product from cart")
				return err
			}
		}

		err = dbHelper.UpdateCartToInactive(tx, cartId, model.CartStatusInActive)

		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, err, "Failed to update status of cart")
			return err
		}

		return nil
	})
	// error message correct karo
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "transaction error")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string
	}{Message: "Order placed successfully"})
}

func CreateOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderId := chi.URLParam(r, "orderId")
	status := chi.URLParam(r, "orderStatus")
	var orderStatus model.OrderStatus
	if status == "shipping" {
		orderStatus = model.OrderStatusShipping
	} else if status == "delivered" {
		orderStatus = model.OrderStatusDelivered
	}
	////var body model.PlacedOrderStatus
	//var status string
	//if err := utils.ParseBody(r.Body, &status); err != nil {
	//	utils.RespondError(w, http.StatusBadRequest, err, "Failed to parse request body")
	//	return
	//}
	//parseBody := model.PlacedOrderStatus{
	//	Status: body.Status,
	//}
	//validate := validator.New()
	//if err := validate.Struct(parseBody); err != nil {
	//	utils.RespondError(w, http.StatusBadRequest, err, "input field is invalid")
	//	return
	//}

	err := dbHelper.CreateOrderStatus(database.Audiophile, orderId, orderStatus)

	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, err, "Failed to create order status")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, struct {
		Message string
	}{"order status changed"})
}

func GetUserAddress(w http.ResponseWriter, r *http.Request) {
	userId := getUserId(r)
	userAddress, err := dbHelper.GetAddress(database.Audiophile, userId)
	logrus.Println(userAddress)
	if err != nil {
		return
	}
	err = utils.EncodeJSONBody(w, userAddress)
	if err != nil {
		return
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
