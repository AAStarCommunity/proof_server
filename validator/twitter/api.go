package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/samber/lo"
	"golang.org/x/xerrors"
)

type APIResponse struct {
	User struct {
		ID         string `json:"user_id"`
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Text string `json:"text"`
}

const (
	BASIC_AUTH_USERNAME = "3rJOl1ODzm9yZy63FACdg"
	BASIC_AUTH_PASSWORD = "5jPoQ5kQvMJFDYRNE8bQ4rHuds4xJqhvgNJM4awaE8"
)

var (
	// TODO: should save accessToken to somewhere else (shared by all Lambda instances)
	accessToken string
	guestToken  string
	flowToken   string
)

func fetchPostWithAPI(id string, maxRetries int) (tweet *APIResponse, err error) {
	const RETRY_AFTER = time.Second

	return nil, nil
}

func setHeaders(req *http.Request, setAccessToken, setGuestToken bool) {
	req.Header.Set("User-Agent", "TwitterAndroid/9.95.0-release.0 (29950000-r-0) ONEPLUS+A3010/9 (OnePlus;ONEPLUS+A3010;OnePlus;OnePlus3;0;;1;2016)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Twitter-API-Version", "5")
	req.Header.Set("X-Twitter-Client", "TwitterAndroid")
	req.Header.Set("X-Twitter-Client-Version", "9.95.0-release.0")
	req.Header.Set("OS-Version", "28")
	req.Header.Set("System-User-Agent", "Dalvik/2.1.0 (Linux; U; Android 9; ONEPLUS A3010 Build/PKQ1.181203.001)")
	req.Header.Set("X-Twitter-Active-User", "yes")
	if setGuestToken {
		req.Header.Set("X-Guest-Token", guestToken)
	}
	if setAccessToken {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
}

func GetOauthToken() (err error) {
	if flowToken == "" {
		if err := getFlowToken(); err != nil {
			return err
		}
	}
	l.Infof("Access token: %s\nGuest token: %s\nFlow token: %s\n", accessToken, guestToken, flowToken)

	requestBody := fmt.Sprintf(`{
        "flow_token": "%s",
        "subtask_inputs": [
            {
                "open_link": {
                    "link": "next_link"
                },
                "subtask_id": "NextTaskOpenLink"
            }
        ],
        "subtask_versions": {
            "generic_urt": 3,
            "standard": 1,
            "open_home_timeline": 1,
            "app_locale_update": 1,
            "enter_date": 1,
            "email_verification": 3,
            "enter_password": 5,
            "enter_text": 5,
            "one_tap": 2,
            "cta": 7,
            "single_sign_on": 1,
            "fetch_persisted_data": 1,
            "enter_username": 3,
            "web_modal": 2,
            "fetch_temporary_password": 1,
            "menu_dialog": 1,
            "sign_up_review": 5,
            "interest_picker": 4,
            "user_recommendations_urt": 3,
            "in_app_notification": 1,
            "sign_up": 2,
            "typeahead_search": 1,
            "user_recommendations_list": 4,
            "cta_inline": 1,
            "contacts_live_sync_permission_prompt": 3,
            "choice_selection": 5,
            "js_instrumentation": 1,
            "alert_dialog_suppress_client_events": 1,
            "privacy_options": 1,
            "topics_selector": 1,
            "wait_spinner": 3,
            "tweet_selection_urt": 1,
            "end_flow": 1,
            "settings_list": 7,
            "open_external_link": 1,
            "phone_verification": 5,
            "security_key": 3,
            "select_banner": 2,
            "upload_media": 1,
            "web": 2,
            "alert_dialog": 1,
            "open_account": 2,
            "action_list": 2,
            "enter_phone": 2,
            "open_link": 1,
            "show_code": 1,
            "update_users": 1,
            "check_logged_in_account": 1,
            "enter_email": 2,
            "select_avatar": 4,
            "location_permission_prompt": 2,
            "notifications_permission_prompt": 4
        }
    }`, flowToken)

	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/onboarding/task.json", strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	setHeaders(req, true, true)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	l.Infof("Response: \n%s\n", body)
	type ResponseSubtask struct {
		// Should exist
		OpenAccount *struct {
			OauthToken       string `json:"oauth_token"`
			OauthTokenSecret string `json:"oauth_token_secret"`
		} `json:"open_account"`
	}

	type Response struct {
		// Should be empty
		Errors *[]struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		// Should be "success"
		Status string `json:"status"`
		// A new flow token, usually ends with ":3"
		FlowToken string             `json:"flow_token"`
		Subtasks  []ResponseSubtask `json:"subtasks"`
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}
	if response.Errors != nil {
		return xerrors.Errorf("error when getting oauth token: %+v", *response.Errors)
	}
	if response.Status != "success" {
		return xerrors.Errorf("wrong API status: %s", response.Status)
	}

	st, found := lo.Find(response.Subtasks, func(subtask ResponseSubtask) bool {
		return (subtask.OpenAccount != nil)
	})
	if !found {
		return xerrors.Errorf("oauth token not found in response")
	}
	flowToken = response.FlowToken
	l.Infof("OAUTH TOKEN REGISTERED: %s --- %s", st.OpenAccount.OauthToken, st.OpenAccount.OauthTokenSecret)

	return nil
}

func getFlowToken() (err error) {
	if guestToken == "" {
		if err := getGuestToken(); err != nil {
			return err
		}
	}

	requestBody := `{
            "flow_token": null,
            "input_flow_data": {
                "country_code": null,
                "flow_context": {
                    "start_location": {
                        "location": "splash_screen"
                    }
                },
                "requested_variant": null,
                "target_user_id": 0
            },
            "subtask_versions": {
                "generic_urt": 3,
                "standard": 1,
                "open_home_timeline": 1,
                "app_locale_update": 1,
                "enter_date": 1,
                "email_verification": 3,
                "enter_password": 5,
                "enter_text": 5,
                "one_tap": 2,
                "cta": 7,
                "single_sign_on": 1,
                "fetch_persisted_data": 1,
                "enter_username": 3,
                "web_modal": 2,
                "fetch_temporary_password": 1,
                "menu_dialog": 1,
                "sign_up_review": 5,
                "interest_picker": 4,
                "user_recommendations_urt": 3,
                "in_app_notification": 1,
                "sign_up": 2,
                "typeahead_search": 1,
                "user_recommendations_list": 4,
                "cta_inline": 1,
                "contacts_live_sync_permission_prompt": 3,
                "choice_selection": 5,
                "js_instrumentation": 1,
                "alert_dialog_suppress_client_events": 1,
                "privacy_options": 1,
                "topics_selector": 1,
                "wait_spinner": 3,
                "tweet_selection_urt": 1,
                "end_flow": 1,
                "settings_list": 7,
                "open_external_link": 1,
                "phone_verification": 5,
                "security_key": 3,
                "select_banner": 2,
                "upload_media": 1,
                "web": 2,
                "alert_dialog": 1,
                "open_account": 2,
                "action_list": 2,
                "enter_phone": 2,
                "open_link": 1,
                "show_code": 1,
                "update_users": 1,
                "check_logged_in_account": 1,
                "enter_email": 2,
                "select_avatar": 4,
                "location_permission_prompt": 2,
                "notifications_permission_prompt": 4
            }
        }`

	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/onboarding/task.json?flow_name=welcome&api_version=1&known_device_token=&sim_country_code=us", strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	setHeaders(req, true, true)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	type Response struct {
		FlowToken string `json:"flow_token"`
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}

	if response.FlowToken == "" {
		return xerrors.Errorf("empty FlowToken")
	}

	flowToken = response.FlowToken
	return nil
}

func getGuestToken() (err error) {
	if guestToken != "" {
		return nil
	}
	if accessToken == "" {
		if err = getAccessToken(); err != nil {
			return err
		}
	}
	req, err := http.NewRequest("POST", "https://api.twitter.com/1.1/guest/activate.json", nil)
	if err != nil {
		return err
	}
	setHeaders(req, true, false)
	type Response struct {
		GuestToken string `json:"guest_token"`
	}

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}

	// Fetching bearerToken
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}
	if response.GuestToken == "" {
		return xerrors.Errorf("Wrong guest token: %s", response.GuestToken)
	}
	guestToken = response.GuestToken

	return nil
}

func getAccessToken() (err error) {
	if accessToken != "" {
		return nil
	}

	type Response struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	req, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/token?grant_type=client_credentials", nil)
	if err != nil {
		return err
	}
	setHeaders(req, false, false)
	req.SetBasicAuth(BASIC_AUTH_USERNAME, BASIC_AUTH_PASSWORD)
	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}

	// Fetching bearerToken
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return err
	}

	if response.TokenType != "bearer" || len(response.AccessToken) == 0 {
		return xerrors.Errorf("Wrong bearer token: %s %s", response.TokenType, response.AccessToken)
	}

	accessToken = response.AccessToken
	return nil
}
