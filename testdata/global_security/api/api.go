package api

import (
	"net/http"
)

// @Summary default security
// @Success 200
// @Router /testapi/application [get]
func GetApplication(w http.ResponseWriter, r *http.Request) {}

// @Summary no security
// @Security
// @Success 200
// @Router /testapi/nosec [get]
func GetNoSec(w http.ResponseWriter, r *http.Request) {}

// @Summary basic security
// @Security BasicAuth
// @Success 200
// @Router /testapi/basic [get]
func GetBasic(w http.ResponseWriter, r *http.Request) {}

// @Summary oauth2 write
// @Security OAuth2Application[write]
// @Success 200
// @Router /testapi/oauth/write [get]
func GetOAuthWrite(w http.ResponseWriter, r *http.Request) {}

// @Summary oauth2 admin
// @Security OAuth2Application[admin]
// @Success 200
// @Router /testapi/oauth/admin [get]
func GetOAuthAdmin(w http.ResponseWriter, r *http.Request) {}
