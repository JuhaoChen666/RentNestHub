package httpapi

import (
	"mime/multipart"
	"net/url"
	"strings"
	"testing"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

func TestHouseFromFormAcceptsValidPayload(t *testing.T) {
	house, err := houseFromForm(validHouseForm())
	if err != nil {
		t.Fatalf("expected valid payload, got %v", err)
	}
	if house.Title != "徐汇滨江明亮两居" {
		t.Fatalf("unexpected title %q", house.Title)
	}
	if len(house.Amenities) != 3 {
		t.Fatalf("expected amenities to be split, got %#v", house.Amenities)
	}
}

func TestHouseFromFormRejectsMissingFields(t *testing.T) {
	form := validHouseForm()
	delete(form.Value, "title")
	delete(form.Value, "monthlyRent")

	_, err := houseFromForm(form)
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	if !strings.Contains(message, "title is required") ||
		!strings.Contains(message, "monthlyRent is required") {
		t.Fatalf("unexpected validation error: %s", message)
	}
}

func TestHouseFromFormRejectsOutOfRangeFields(t *testing.T) {
	form := validHouseForm()
	form.Value["monthlyRent"] = []string{"0"}
	form.Value["bedrooms"] = []string{"25"}
	form.Value["areaSqm"] = []string{"3000"}

	_, err := houseFromForm(form)
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, expected := range []string{
		"monthlyRent must be between 1 and 200000",
		"bedrooms must be between 1 and 20",
		"areaSqm must be between 1 and 2000",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected %q in %q", expected, message)
		}
	}
}

func TestDetectImageExtensionAcceptsSupportedImages(t *testing.T) {
	cases := []struct {
		name      string
		sample    []byte
		extension string
	}{
		{
			name:      "jpeg",
			sample:    []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00},
			extension: ".jpg",
		},
		{
			name:      "png",
			sample:    []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x00, 0x00, 0x00, '\r', 'I', 'H', 'D', 'R'},
			extension: ".png",
		},
		{
			name:      "webp",
			sample:    []byte{'R', 'I', 'F', 'F', 0x1a, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P', 'V', 'P', '8', ' '},
			extension: ".webp",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			extension, ok := detectImageExtension(testCase.sample)
			if !ok || extension != testCase.extension {
				t.Fatalf("expected %s, got %q accepted=%v", testCase.extension, extension, ok)
			}
		})
	}
}

func TestDetectImageExtensionRejectsNonImages(t *testing.T) {
	if extension, ok := detectImageExtension([]byte("plain text")); ok {
		t.Fatalf("expected non-image to be rejected, got %q", extension)
	}
}

func TestHouseFilterFromQueryAcceptsValidQuery(t *testing.T) {
	filter, err := houseFilterFromQuery(url.Values{
		"city":     {" 上海 "},
		"district": {"徐汇区"},
		"keyword":  {"近地铁"},
		"minRent":  {"3000"},
		"maxRent":  {"7000"},
		"bedrooms": {"2"},
		"limit":    {"20"},
		"offset":   {"40"},
		"sort":     {"rent_desc"},
	})
	if err != nil {
		t.Fatalf("expected valid query, got %v", err)
	}
	if filter.City != "上海" || filter.MinRent != 3000 ||
		filter.MaxRent != 7000 || filter.Limit != 20 || filter.Sort != "rent_desc" || !filter.OnlyActive {
		t.Fatalf("unexpected filter: %#v", filter)
	}
}

func TestHouseFilterFromQueryRejectsInvalidNumbers(t *testing.T) {
	_, err := houseFilterFromQuery(url.Values{
		"maxRent": {"free"},
		"limit":   {"500"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, expected := range []string{
		"maxRent must be between 1 and 200000",
		"limit must be between 1 and 100",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("expected %q in %q", expected, message)
		}
	}
}

func TestHouseFilterFromQueryRejectsInvertedRentRange(t *testing.T) {
	_, err := houseFilterFromQuery(url.Values{
		"minRent": {"8000"},
		"maxRent": {"5000"},
	})
	if err == nil || !strings.Contains(err.Error(), "minRent cannot exceed maxRent") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHouseFilterFromQueryUsesDefaultLimit(t *testing.T) {
	filter, err := houseFilterFromQuery(url.Values{})
	if err != nil {
		t.Fatalf("expected empty query to be valid, got %v", err)
	}
	if filter.Limit != defaultSearchLimit || filter.Offset != 0 {
		t.Fatalf("unexpected pagination defaults: %#v", filter)
	}
	if filter.Sort != searchSortLatest {
		t.Fatalf("unexpected sort default: %#v", filter)
	}
}

func TestNewListHousesResponseIncludesPaginationMeta(t *testing.T) {
	response := newListHousesResponse(
		[]domain.House{{ID: 1}, {ID: 2}},
		domain.HouseFilter{Limit: 2, Offset: 4, Sort: "rent_asc"},
	)
	if response.Meta.Count != 2 || response.Meta.Offset != 4 || !response.Meta.HasMore || response.Meta.Sort != "rent_asc" {
		t.Fatalf("unexpected pagination meta: %#v", response.Meta)
	}
}

func TestHouseFilterFromQueryRejectsInvalidSort(t *testing.T) {
	_, err := houseFilterFromQuery(url.Values{"sort": {"popular"}})
	if err == nil || !strings.Contains(err.Error(), "sort must be one of latest, rent_asc, rent_desc") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func validHouseForm() *multipart.Form {
	return &multipart.Form{Value: map[string][]string{
		"landlordId":  {"1"},
		"title":       {"徐汇滨江明亮两居"},
		"description": {"朝南客厅，步行可达地铁站，适合两人合租或小家庭。"},
		"city":        {"上海"},
		"district":    {"徐汇区"},
		"address":     {"龙腾大道附近"},
		"monthlyRent": {"6200"},
		"bedrooms":    {"2"},
		"bathrooms":   {"1"},
		"areaSqm":     {"68"},
		"amenities":   {"近地铁, 电梯, 独立阳台"},
	}}
}
