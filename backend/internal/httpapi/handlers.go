package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/repository/mysqlrepo"
)

const maxUploadSize = 20 << 20

func (api *API) listHouses(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	houses, err := api.repository.ListHouses(request.Context(), domain.HouseFilter{
		City:       strings.TrimSpace(query.Get("city")),
		District:   strings.TrimSpace(query.Get("district")),
		Keyword:    strings.TrimSpace(query.Get("keyword")),
		MinRent:    intQuery(query.Get("minRent")),
		MaxRent:    intQuery(query.Get("maxRent")),
		Bedrooms:   intQuery(query.Get("bedrooms")),
		Limit:      intQuery(query.Get("limit")),
		Offset:     intQuery(query.Get("offset")),
		OnlyActive: true,
	})
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": houses})
}

func (api *API) getHouse(writer http.ResponseWriter, request *http.Request) {
	id, err := strconv.ParseInt(request.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(writer, http.StatusBadRequest, "invalid house id")
		return
	}
	house, err := api.repository.GetHouse(request.Context(), id)
	if errors.Is(err, mysqlrepo.ErrNotFound) {
		writeError(writer, http.StatusNotFound, "house not found")
		return
	}
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, house)
}

func (api *API) createHouse(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxUploadSize)
	if err := request.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid form or upload exceeds 20 MB")
		return
	}

	house, err := houseFromForm(request.MultipartForm)
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	house.ImageURLs, err = api.saveImages(request.MultipartForm.File["images"])
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if err := api.repository.CreateHouse(request.Context(), &house); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusCreated, house)
}

func (api *API) recommend(writer http.ResponseWriter, request *http.Request) {
	var input domain.RecommendationRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(input.Need) == "" {
		writeError(writer, http.StatusBadRequest, "need is required")
		return
	}
	result, err := api.recommender.Recommend(request.Context(), input)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"items": result,
		"mode":  "local-ranking",
	})
}

func (api *API) addFavorite(writer http.ResponseWriter, request *http.Request) {
	var favorite domain.Favorite
	if err := decodeJSON(request, &favorite); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	if favorite.TenantID <= 0 || favorite.HouseID <= 0 {
		writeError(writer, http.StatusBadRequest, "tenantId and houseId are required")
		return
	}
	if err := api.repository.AddFavorite(request.Context(), favorite); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (api *API) removeFavorite(writer http.ResponseWriter, request *http.Request) {
	tenantID, tenantErr := strconv.ParseInt(request.PathValue("tenantId"), 10, 64)
	houseID, houseErr := strconv.ParseInt(request.PathValue("houseId"), 10, 64)
	if tenantErr != nil || houseErr != nil || tenantID <= 0 || houseID <= 0 {
		writeError(writer, http.StatusBadRequest, "invalid tenant or house id")
		return
	}
	if err := api.repository.RemoveFavorite(request.Context(), domain.Favorite{
		TenantID: tenantID,
		HouseID:  houseID,
	}); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (api *API) createMessage(writer http.ResponseWriter, request *http.Request) {
	var message domain.Message
	if err := decodeJSON(request, &message); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	message.Content = strings.TrimSpace(message.Content)
	if message.HouseID <= 0 || message.SenderID <= 0 || message.Content == "" {
		writeError(writer, http.StatusBadRequest, "houseId, senderId, and content are required")
		return
	}
	if len([]rune(message.Content)) > 1000 {
		writeError(writer, http.StatusBadRequest, "message cannot exceed 1000 characters")
		return
	}
	if err := api.repository.CreateMessage(request.Context(), &message); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusCreated, message)
}

func houseFromForm(form *multipart.Form) (domain.House, error) {
	value := form.Value
	landlordID, _ := strconv.ParseInt(first(value["landlordId"]), 10, 64)
	monthlyRent, _ := strconv.Atoi(first(value["monthlyRent"]))
	bedrooms, _ := strconv.Atoi(first(value["bedrooms"]))
	bathrooms, _ := strconv.Atoi(first(value["bathrooms"]))
	areaSqm, _ := strconv.ParseFloat(first(value["areaSqm"]), 64)

	house := domain.House{
		LandlordID:  landlordID,
		Title:       strings.TrimSpace(first(value["title"])),
		Description: strings.TrimSpace(first(value["description"])),
		City:        strings.TrimSpace(first(value["city"])),
		District:    strings.TrimSpace(first(value["district"])),
		Address:     strings.TrimSpace(first(value["address"])),
		MonthlyRent: monthlyRent,
		Bedrooms:    bedrooms,
		Bathrooms:   bathrooms,
		AreaSqm:     areaSqm,
		Amenities:   splitCSV(first(value["amenities"])),
		Status:      "active",
	}
	if house.LandlordID <= 0 || house.Title == "" || house.City == "" ||
		house.District == "" || house.MonthlyRent <= 0 || house.Bedrooms <= 0 {
		return domain.House{}, errors.New(
			"landlordId, title, city, district, monthlyRent, and bedrooms are required",
		)
	}
	return house, nil
}

func (api *API) saveImages(files []*multipart.FileHeader) ([]string, error) {
	if len(files) > 8 {
		return nil, errors.New("a house can contain at most 8 images")
	}
	if err := os.MkdirAll(api.uploadDir, 0o755); err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(files))
	for index, header := range files {
		contentType := header.Header.Get("Content-Type")
		extension := map[string]string{
			"image/jpeg": ".jpg",
			"image/png":  ".png",
			"image/webp": ".webp",
		}[contentType]
		if extension == "" {
			return nil, fmt.Errorf("unsupported image type: %s", contentType)
		}

		source, err := header.Open()
		if err != nil {
			return nil, err
		}
		filename := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), index, extension)
		target, err := os.OpenFile(
			filepath.Join(api.uploadDir, filename),
			os.O_WRONLY|os.O_CREATE|os.O_EXCL,
			0o644,
		)
		if err != nil {
			source.Close()
			return nil, err
		}
		_, copyErr := io.Copy(target, source)
		source.Close()
		target.Close()
		if copyErr != nil {
			return nil, copyErr
		}
		urls = append(urls, strings.TrimRight(api.publicBaseURL, "/")+"/uploads/"+filename)
	}
	return urls, nil
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	return nil
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func intQuery(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}
