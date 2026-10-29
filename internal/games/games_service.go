// for bussiness logic
package games

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func CallSteamSearch(ctx context.Context, game string) (SteamSearchResponse, error) {
	var searchResp SteamSearchResponse
	baseURL := "https://store.steampowered.com/api/storesearch/"
	params := url.Values{}
	params.Add("term", game)
	params.Add("cc", "ar")
	params.Add("l", "es")

	fullURL := baseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return SteamSearchResponse{}, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SteamSearchResponse{}, fmt.Errorf("calling steam API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SteamSearchResponse{}, fmt.Errorf("steam API returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return SteamSearchResponse{}, fmt.Errorf("error during decode: %v", err)
	}
	return searchResp, nil
}

func CallSteamSearchByID(ctx context.Context, steamAppID int) (SteamAppDetails, error) {
	baseURL := "https://store.steampowered.com/api/appdetails"
	params := url.Values{}
	params.Add("appids", strconv.Itoa(steamAppID))
	params.Add("cc", "ar")
	params.Add("l", "es")
	
	fullURL := baseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return SteamAppDetails{}, fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return SteamAppDetails{}, fmt.Errorf("calling steam API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SteamAppDetails{}, fmt.Errorf("steam API returned status %d", resp.StatusCode)
	}	

	var steamResp map[string]SteamAppDetails
	if err := json.NewDecoder(resp.Body).Decode(&steamResp); err != nil {
		return SteamAppDetails{}, err
	}
	
	result := steamResp[strconv.Itoa(steamAppID)]
	if result.Success && result.Data != nil && result.Data.PriceOverview != nil {
		result.Data.PriceOverview.Final = result.Data.PriceOverview.Final / 100
	}
	return result, nil
}

func GameDetailsLogic(steamResp SteamAppDetails, steamAppID int, targetPrice float64) (GameResponse, error) {
	if steamResp.Success == false || steamResp.Data == nil {
		return GameResponse{}, errors.New("Steam data is nil")
	} 

	currentPrice := 0.0
	if steamResp.Data.PriceOverview != nil {
 		currentPrice = float64(steamResp.Data.PriceOverview.Final)
	}
	
	if !steamResp.Data.IsFree && steamResp.Data.PriceOverview != nil && targetPrice == 0.0 {
		targetPrice = float64(steamResp.Data.PriceOverview.Final) / 100
	}

	return GameResponse{
		SteamAppID: steamAppID,
		Name: steamResp.Data.Name,
		CurrentPrice: currentPrice,
		TargetPrice: targetPrice,
		IsFree: steamResp.Data.IsFree,
	}, nil
}

// obtener el user id con el jwt (esto quizas sea mejor ponerlo en otro lado)
func GetTokenClaims(ctx context.Context) (uuid.UUID, error) {
	claims, ok := ctx.Value("claims").(jwt.MapClaims)
	if !ok {
    	return uuid.Nil, errors.New("Invalid Token")
	}
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("Invalid Token")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("Invalid Token")
	}
	return userID, nil
}

