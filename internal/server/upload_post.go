package server

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"

	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/gojwk"
	"github.com/gin-gonic/gin"
	"github.com/liteseed/transit/internal/database/schema"
)

func readDataFromMultipartFile(context *gin.Context) ([]byte, error) {

	fileHeader, err := context.FormFile("file")
	if err != nil {
		context.Error(err)
		return nil, errors.New("multipart/form-data with key file expected")
	}

	data := make([]byte, fileHeader.Size)
	multipartFile, err := fileHeader.Open()
	if err != nil {
		context.Error(err)
		return nil, errors.New("unable to read file")
	}

	_, err = multipartFile.Read(data)
	if err != nil {
		context.Error(err)
		return nil, errors.New("unable to read file")
	}
	return data, nil
}

func generateNewSigner() (*goar.ItemSigner, error) {
	bitSize := 4096

	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}
	jwk, err := gojwk.PrivateKey(key)
	if err != nil {
		return nil, err
	}
	data, err := gojwk.Marshal(jwk)
	if err != nil {
		return nil, err
	}
	signer, err := goar.NewSigner(data)
	if err != nil {
		return nil, err
	}

	itemSigner, err := goar.NewItemSigner(signer)
	if err != nil {
		return nil, err
	}

	return itemSigner, err
}

// POST /upload
// Basic Steps
// 1. Check Gateway is available
// 2. Check if the transaction pays enough AR (TODO)
// 3. readDataFromMultipartFile - Read multipart/form-data file and get bytes
// 4. Generate a new signer to sign the file
// 5. Create a data-item
// 6. Assign a bundler
// 7. Create an order
// 8. 

func (s *Server) UploadPost(context *gin.Context) {
	info, err := s.wallet.Client.GetInfo()
	if err != nil {
		context.JSON(http.StatusFailedDependency, gin.H{"error": "failed to query gateway"})
		context.Error(err)
		return
	}

	data, err := readDataFromMultipartFile(context)
	if err != nil {
		context.JSON(http.StatusBadRequest, err)
		return
	}

	signer, err := generateNewSigner()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create signer"})
		context.Error(err)
		return
	}

	dataItem, err := signer.CreateAndSignItem(data, "", "", []types.Tag{{}})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create data-item"})
		context.Error(err)
		return
	}

	o := &schema.Order{
		ID:      dataItem.Id,
		Status:  schema.Queued,
		Price:   uint64(0),
		Payment: schema.Paid,
	}

	err = s.store.Set(dataItem.Id, dataItem.ItemBinary)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		context.Error(err)
		return
	}
	err = s.database.CreateOrder(o)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		context.Error(err)
		return
	}

	context.JSON(
		http.StatusCreated,
		&DataItemPostResponse{
			ID:                  o.ID,
			Owner:               s.wallet.Signer.Address,
			Price:               o.Price,
			Version:             "1.0.0",
			DeadlineHeight:      uint(info.Height + 200),
			DataCaches:          []string{s.gateway},
			FastFinalityIndexes: []string{s.gateway},
		})
}
