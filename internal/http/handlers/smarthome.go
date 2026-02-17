package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/services/homeassistant"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

type SmartHomeHandler struct {
	svc services.ISmartHomeService
}

func NewSmartHomeHandler(svc services.ISmartHomeService) *SmartHomeHandler {
	return &SmartHomeHandler{svc: svc}
}

// Connect godoc
// @Summary      Connect Home Assistant
// @Description  Connect a Home Assistant server to the home
// @Tags         smarthome
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        request body models.ConnectHARequest true "HA Connection"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/connect [post]
func (h *SmartHomeHandler) Connect(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.ConnectHARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	if err := h.svc.ConnectHA(r.Context(), homeID, req.URL, req.Token); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Home Assistant connected successfully",
	})
}

// Disconnect godoc
// @Summary      Disconnect Home Assistant
// @Description  Disconnect Home Assistant from the home
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/disconnect [delete]
func (h *SmartHomeHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.DisconnectHA(r.Context(), homeID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Home Assistant disconnected",
	})
}

// Status godoc
// @Summary      Get Home Assistant status
// @Description  Get connection status and config
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/status [get]
func (h *SmartHomeHandler) Status(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	config, err := h.svc.GetHAConfig(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if config == nil {
		utils.JSON(w, http.StatusOK, map[string]interface{}{
			"connected": false,
		})
		return
	}

	// Test connection
	connErr := h.svc.TestConnection(r.Context(), homeID)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"connected":  connErr == nil,
		"url":        config.URL,
		"is_active":  config.IsActive,
		"created_at": config.CreatedAt,
		"error":      errorToString(connErr),
	})
}

// Discover godoc
// @Summary      Discover devices
// @Description  Discover all available devices from Home Assistant
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  []homeassistant.HAState
// @Router       /homes/{home_id}/smarthome/discover [get]
func (h *SmartHomeHandler) Discover(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	devices, err := h.svc.DiscoverDevices(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"devices": devices,
		"count":   len(devices),
	})
}

// AddDevice godoc
// @Summary      Add device
// @Description  Add a device from Home Assistant to the app
// @Tags         smarthome
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        request body models.AddDeviceRequest true "Device Info"
// @Success      201  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/devices [post]
func (h *SmartHomeHandler) AddDevice(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	var req models.AddDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	deviceType := homeassistant.GetEntityDomain(req.EntityID)
	if err := h.svc.AddDevice(r.Context(), homeID, req.EntityID, req.Name, deviceType, req.RoomID, req.Icon); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"message": "Device added successfully",
	})
}

// GetDevices godoc
// @Summary      Get devices
// @Description  Get all added devices for the home
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  []models.SmartDevice
// @Router       /homes/{home_id}/smarthome/devices [get]
func (h *SmartHomeHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	devices, err := h.svc.GetDevices(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"devices": devices,
	})
}

// GetDevice godoc
// @Summary      Get device
// @Description  Get device details with current state
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        device_id path int true "Device ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/devices/{device_id} [get]
func (h *SmartHomeHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	deviceIDStr := chi.URLParam(r, "device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	device, err := h.svc.GetDeviceByID(r.Context(), deviceID, homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if device == nil {
		utils.JSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	// Get current state from HA
	state, stateErr := h.svc.GetDeviceState(r.Context(), homeID, device.EntityID)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"device": device,
		"state":  state,
		"error":  errorToString(stateErr),
	})
}

// UpdateDevice godoc
// @Summary      Update device
// @Description  Update device name, room, or icon
// @Tags         smarthome
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        device_id path int true "Device ID"
// @Param        request body models.UpdateDeviceRequest true "Device Update"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/devices/{device_id} [put]
func (h *SmartHomeHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {

	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	deviceIDStr := chi.URLParam(r, "device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	if err := h.svc.UpdateDevice(r.Context(), deviceID, homeID, req.Name, req.RoomID, req.Icon); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Device updated successfully",
	})
}

// DeleteDevice godoc
// @Summary      Delete device
// @Description  Remove device from the app
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        device_id path int true "Device ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/devices/{device_id} [delete]
func (h *SmartHomeHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	deviceIDStr := chi.URLParam(r, "device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	if err := h.svc.RemoveDevice(r.Context(), deviceID); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Device removed successfully",
	})
}

// ControlDevice godoc
// @Summary      Control device
// @Description  Call a service on the device (turn_on, turn_off, etc.)
// @Tags         smarthome
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        device_id path int true "Device ID"
// @Param        request body models.ControlDeviceRequest true "Control Command"
// @Success      200  {object}  map[string]interface{}
// @Router       /homes/{home_id}/smarthome/devices/{device_id}/control [post]
func (h *SmartHomeHandler) ControlDevice(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	deviceIDStr := chi.URLParam(r, "device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid device ID", http.StatusBadRequest)
		return
	}

	var req models.ControlDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(req); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	// Get device to get entity_id
	device, err := h.svc.GetDeviceByID(r.Context(), deviceID, homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if device == nil {
		utils.JSONError(w, "Device not found", http.StatusNotFound)
		return
	}

	if err := h.svc.ControlDevice(r.Context(), homeID, device.EntityID, req.Service, req.Data); err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Command executed successfully",
	})
}

// GetAllStates godoc
// @Summary      Get all device states
// @Description  Get current states of all added devices
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Success      200  {object}  []homeassistant.HAState
// @Router       /homes/{home_id}/smarthome/states [get]
func (h *SmartHomeHandler) GetAllStates(w http.ResponseWriter, r *http.Request) {
	homeIDStr := chi.URLParam(r, "home_id")
	homeID, err := strconv.Atoi(homeIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid home ID", http.StatusBadRequest)
		return
	}

	states, err := h.svc.GetAllStates(r.Context(), homeID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"states": states,
	})
}

// GetDevicesByRoom godoc
// @Summary      Get devices by room
// @Description  Get all devices assigned to a room
// @Tags         smarthome
// @Produce      json
// @Security     BearerAuth
// @Param        home_id path int true "Home ID"
// @Param        room_id path int true "Room ID"
// @Success      200  {object}  []models.SmartDevice
// @Router       /homes/{home_id}/rooms/{room_id}/devices [get]
func (h *SmartHomeHandler) GetDevicesByRoom(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		utils.JSONError(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	devices, err := h.svc.GetDevicesByRoom(r.Context(), roomID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"devices": devices,
	})
}

// Helper function
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
