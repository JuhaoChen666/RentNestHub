package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/repository/mysqlrepo"
)

const (
	maxUploadSize = 20 << 20
	maxImageFiles = 8
	maxImageSize  = 5 << 20
	sniffSize     = 512

	defaultSearchLimit = 24
	maxSearchLimit     = 100
	maxSearchOffset    = 10000
	searchSortLatest   = "latest"
	searchSortRentAsc  = "rent_asc"
	searchSortRentDesc = "rent_desc"
)

type listHousesResponse struct {
	Items []domain.House `json:"items"`
	Meta  listHousesMeta `json:"meta"`
}

type listHousesMeta struct {
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	Count   int    `json:"count"`
	HasMore bool   `json:"hasMore"`
	Sort    string `json:"sort"`
}

func (api *API) listHouses(writer http.ResponseWriter, request *http.Request) {
	filter, err := houseFilterFromQuery(request.URL.Query())
	if err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	houses, err := api.repository.ListHouses(request.Context(), filter)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, newListHousesResponse(houses, filter))
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
		"mode":  api.recommender.Mode(),
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

func (api *API) listMessages(writer http.ResponseWriter, request *http.Request) {
	senderID, err := strconv.ParseInt(request.PathValue("senderId"), 10, 64)
	if err != nil || senderID <= 0 {
		writeError(writer, http.StatusBadRequest, "invalid sender id")
		return
	}
	messages, err := api.repository.ListMessages(request.Context(), senderID)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": messages})
}

func houseFromForm(form *multipart.Form) (domain.House, error) {
	value := form.Value
	landlordID, landlordErr := parsePositiveInt64(value, "landlordId")
	monthlyRent, rentErr := parseBoundedInt(value, "monthlyRent", 1, 200000)
	bedrooms, bedroomsErr := parseBoundedInt(value, "bedrooms", 1, 20)
	bathrooms, bathroomsErr := parseBoundedInt(value, "bathrooms", 1, 20)
	areaSqm, areaErr := parseBoundedFloat(value, "areaSqm", 1, 2000)

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

	var validationErrors []string
	validationErrors = append(validationErrors, landlordErr...)
	validationErrors = append(validationErrors, rentErr...)
	validationErrors = append(validationErrors, bedroomsErr...)
	validationErrors = append(validationErrors, bathroomsErr...)
	validationErrors = append(validationErrors, areaErr...)
	validationErrors = append(validationErrors, validateText("title", house.Title, 1, 120)...)
	validationErrors = append(validationErrors, validateText("description", house.Description, 1, 1000)...)
	validationErrors = append(validationErrors, validateText("city", house.City, 1, 80)...)
	validationErrors = append(validationErrors, validateText("district", house.District, 1, 80)...)
	validationErrors = append(validationErrors, validateText("address", house.Address, 1, 180)...)
	validationErrors = append(validationErrors, validateAmenities(house.Amenities)...)
	if len(validationErrors) > 0 {
		return domain.House{}, errors.New("invalid house payload: " + strings.Join(validationErrors, "; "))
	}
	return house, nil
}

func (api *API) saveImages(files []*multipart.FileHeader) ([]string, error) {
	if len(files) > maxImageFiles {
		return nil, fmt.Errorf("a house can contain at most %d images", maxImageFiles)
	}
	if err := os.MkdirAll(api.uploadDir, 0o755); err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(files))
	for index, header := range files {
		if header.Size <= 0 {
			return nil, fmt.Errorf("%s is empty", header.Filename)
		}
		if header.Size > maxImageSize {
			return nil, fmt.Errorf("%s exceeds the 5 MB image limit", header.Filename)
		}

		source, err := header.Open()
		if err != nil {
			return nil, err
		}
		sniff := make([]byte, sniffSize)
		n, readErr := io.ReadFull(source, sniff)
		if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) && !errors.Is(readErr, io.EOF) {
			source.Close()
			return nil, readErr
		}
		if n == 0 {
			source.Close()
			return nil, fmt.Errorf("%s is empty", header.Filename)
		}

		extension, ok := detectImageExtension(sniff[:n])
		if !ok {
			source.Close()
			return nil, fmt.Errorf("%s must be a JPEG, PNG, or WebP image", header.Filename)
		}

		filename := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), index, extension)
		path := filepath.Join(api.uploadDir, filename)
		target, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			source.Close()
			return nil, err
		}
		_, writeErr := target.Write(sniff[:n])
		written, copyErr := io.Copy(target, io.LimitReader(source, maxImageSize-int64(n)+1))
		source.Close()
		target.Close()
		if writeErr != nil {
			_ = os.Remove(path)
			return nil, writeErr
		}
		if copyErr != nil {
			_ = os.Remove(path)
			return nil, copyErr
		}
		if int64(n)+written > maxImageSize {
			_ = os.Remove(path)
			return nil, fmt.Errorf("%s exceeds the 5 MB image limit", header.Filename)
		}
		urls = append(urls, strings.TrimRight(api.publicBaseURL, "/")+"/uploads/"+filename)
	}
	return urls, nil
}

func detectImageExtension(sample []byte) (string, bool) {
	switch http.DetectContentType(sample) {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func houseFilterFromQuery(query url.Values) (domain.HouseFilter, error) {
	minRent, minRentErrors := parseOptionalBoundedInt(query, "minRent", 1, 200000)
	maxRent, maxRentErrors := parseOptionalBoundedInt(query, "maxRent", 1, 200000)
	bedrooms, bedroomErrors := parseOptionalBoundedInt(query, "bedrooms", 1, 20)
	limit, limitErrors := parseOptionalBoundedInt(query, "limit", 1, maxSearchLimit)
	offset, offsetErrors := parseOptionalBoundedInt(query, "offset", 0, maxSearchOffset)

	var validationErrors []string
	validationErrors = append(validationErrors, minRentErrors...)
	validationErrors = append(validationErrors, maxRentErrors...)
	validationErrors = append(validationErrors, bedroomErrors...)
	validationErrors = append(validationErrors, limitErrors...)
	validationErrors = append(validationErrors, offsetErrors...)
	if minRent > 0 && maxRent > 0 && minRent > maxRent {
		validationErrors = append(validationErrors, "minRent cannot exceed maxRent")
	}

	city := strings.TrimSpace(query.Get("city"))
	district := strings.TrimSpace(query.Get("district"))
	keyword := strings.TrimSpace(query.Get("keyword"))
	sort := strings.TrimSpace(query.Get("sort"))
	if sort == "" {
		sort = searchSortLatest
	}
	validationErrors = append(validationErrors, validateOptionalText("city", city, 80)...)
	validationErrors = append(validationErrors, validateOptionalText("district", district, 80)...)
	validationErrors = append(validationErrors, validateOptionalText("keyword", keyword, 80)...)
	validationErrors = append(validationErrors, validateSearchSort(sort)...)
	if len(validationErrors) > 0 {
		return domain.HouseFilter{}, errors.New("invalid search query: " + strings.Join(validationErrors, "; "))
	}
	if limit == 0 {
		limit = defaultSearchLimit
	}

	return domain.HouseFilter{
		City:       city,
		District:   district,
		Keyword:    keyword,
		MinRent:    minRent,
		MaxRent:    maxRent,
		Bedrooms:   bedrooms,
		Limit:      limit,
		Offset:     offset,
		Sort:       sort,
		OnlyActive: true,
	}, nil
}

func newListHousesResponse(houses []domain.House, filter domain.HouseFilter) listHousesResponse {
	return listHousesResponse{
		Items: houses,
		Meta: listHousesMeta{
			Limit:   filter.Limit,
			Offset:  filter.Offset,
			Count:   len(houses),
			HasMore: len(houses) == filter.Limit,
			Sort:    filter.Sort,
		},
	}
}

func validateSearchSort(sort string) []string {
	switch sort {
	case searchSortLatest, searchSortRentAsc, searchSortRentDesc:
		return nil
	default:
		return []string{"sort must be one of latest, rent_asc, rent_desc"}
	}
}

func parseOptionalBoundedInt(
	query url.Values,
	field string,
	minimum int,
	maximum int,
) (int, []string) {
	raw := strings.TrimSpace(query.Get(field))
	if raw == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, []string{fmt.Sprintf("%s must be between %d and %d", field, minimum, maximum)}
	}
	return parsed, nil
}

func validateOptionalText(field string, value string, maximum int) []string {
	if len([]rune(value)) > maximum {
		return []string{fmt.Sprintf("%s cannot exceed %d characters", field, maximum)}
	}
	return nil
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

func parsePositiveInt64(value map[string][]string, field string) (int64, []string) {
	raw := strings.TrimSpace(first(value[field]))
	if raw == "" {
		return 0, []string{field + " is required"}
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, []string{field + " must be a positive integer"}
	}
	return parsed, nil
}

func parseBoundedInt(
	value map[string][]string,
	field string,
	minimum int,
	maximum int,
) (int, []string) {
	raw := strings.TrimSpace(first(value[field]))
	if raw == "" {
		return 0, []string{field + " is required"}
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, []string{fmt.Sprintf("%s must be between %d and %d", field, minimum, maximum)}
	}
	return parsed, nil
}

func parseBoundedFloat(
	value map[string][]string,
	field string,
	minimum float64,
	maximum float64,
) (float64, []string) {
	raw := strings.TrimSpace(first(value[field]))
	if raw == "" {
		return 0, []string{field + " is required"}
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, []string{fmt.Sprintf("%s must be between %.0f and %.0f", field, minimum, maximum)}
	}
	return parsed, nil
}

func validateText(field string, value string, minimum int, maximum int) []string {
	length := len([]rune(strings.TrimSpace(value)))
	if length < minimum {
		return []string{field + " is required"}
	}
	if length > maximum {
		return []string{fmt.Sprintf("%s cannot exceed %d characters", field, maximum)}
	}
	return nil
}

func validateAmenities(amenities []string) []string {
	if len(amenities) > 12 {
		return []string{"amenities cannot contain more than 12 items"}
	}
	for _, amenity := range amenities {
		if len([]rune(amenity)) > 30 {
			return []string{"amenities cannot exceed 30 characters each"}
		}
	}
	return nil
}
