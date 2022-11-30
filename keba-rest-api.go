package main

import (
	modbusclient "github.com/dpapathanasiou/go-modbus"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

type environment struct {
	wallboxName string
	wallboxPort int
	apiPort     string
	hostName    string
	debug       bool
}

type register struct {
	id          int
	description string
	value       int32
}

type phase struct {
	ChargingMAh float64 `json:"charging_mAh"`
	VoltageV    float64 `json:"voltage_V"`
}

// state represents data about a record state.
type state struct {
	ActivePowerMw          float64 `json:"active_power_mW"`
	CableState             int64   `json:"cable_state"`
	ChargedEnergyWh        float64 `json:"charged_energy_Wh"`
	ChargingState          float64 `json:"charging_state"`
	ErrorCode              int64   `json:"error_code"`
	FirmwareVersion        float64 `json:"firmware_version"`
	MaxChargingCurrentMAh  float64 `json:"max_charging_current_mAh"`
	MaxSupportedCurrentMAh float64 `json:"max_supported_current_mAh"`
	PowerFactorPercent     float64 `json:"power_factor_percent"`
	ProductTypeAndFeatures float64 `json:"product_type_and_features"`
	SerialNumber           float64 `json:"serial_number"`
	TotalEnergyCounterWh   float64 `json:"total_energy_counter_Wh"`
	Phase                  []phase `json:"phase"`
	Timestamp              string  `json:"Timestamp"`
}

// getAlbums responds with the list of all currentState as JSON.
func getState(c *gin.Context) {
	log.Debug().Msg("getState")

	for !registerFilled {
		time.Sleep(3 * time.Second)
	}

	currentState.Timestamp = time.Now().Format(time.RFC3339)
	c.IndentedJSON(http.StatusOK, currentState)
}

var (
	registerFilled bool
	registers      []register
	env            environment
	currentState   = state{
		Timestamp: time.Now().Format(time.RFC3339),
		Phase: []phase{
			{ChargingMAh: 0, VoltageV: 0},
			{ChargingMAh: 0, VoltageV: 0},
			{ChargingMAh: 0, VoltageV: 0},
		},
	}
)

func initRegisters() {

	log.Debug().Msg("Init registers")

	registerFilled = false

	registers = append(registers, register{id: 1000, description: "charging_state", value: 0})
	registers = append(registers, register{id: 1004, description: "cable_state", value: 0})
	registers = append(registers, register{id: 1006, description: "error_code", value: 0})
	registers = append(registers, register{id: 1014, description: "serial_number", value: 0})
	registers = append(registers, register{id: 1016, description: "product_type_and_features", value: 0})
	registers = append(registers, register{id: 1018, description: "firmware_version", value: 0})
	registers = append(registers, register{id: 1020, description: "active_power_mW", value: 0})
	registers = append(registers, register{id: 1036, description: "total_energy_counter_Wh", value: 0})
	registers = append(registers, register{id: 1046, description: "power_factor_percent", value: 0})
	registers = append(registers, register{id: 1100, description: "max_charging_current_mAh", value: 0})
	registers = append(registers, register{id: 1110, description: "max_supported_current_mAh", value: 0})
	//registers = append(registers, register{id: 1502 ,description: "rfid_card", value: 0})
	registers = append(registers, register{id: 1502, description: "charged_energy_Wh", value: 0})

	registers = append(registers, register{id: 1008, description: "charging_current_phase_1_mAh", value: 0})
	registers = append(registers, register{id: 1010, description: "charging_current_phase_2_mAh", value: 0})
	registers = append(registers, register{id: 1012, description: "charging_current_phase_3_mAh", value: 0})
	registers = append(registers, register{id: 1040, description: "voltage_phase_1_V", value: 0})
	registers = append(registers, register{id: 1042, description: "voltage_phase_2_V", value: 0})
	registers = append(registers, register{id: 1044, description: "voltage_phase_3_V", value: 0})

}

func main() {

	initApp()

	go func() {
		for {
			updateRegisterData()
			for _, register := range registers {
				switch register.id {
				case 1000:
					currentState.ChargingState = float64(register.value)
				case 1004:
					currentState.CableState = int64(register.value)
				case 1006:
					currentState.ErrorCode = int64(register.value)
				case 1014:
					currentState.SerialNumber = float64(register.value)
				case 1016:
					currentState.ProductTypeAndFeatures = float64(register.value)
				case 1018:
					currentState.FirmwareVersion = float64(register.value)
				case 1020:
					currentState.ActivePowerMw = float64(register.value)
				case 1036:
					currentState.TotalEnergyCounterWh = float64(register.value / 10)
				case 1046:
					currentState.PowerFactorPercent = float64(register.value)
				case 1100:
					currentState.MaxChargingCurrentMAh = float64(register.value)
				case 1110:
					currentState.MaxSupportedCurrentMAh = float64(register.value)
				case 1502:
					currentState.ChargedEnergyWh = float64(register.value / 10)
				case 1008:
					currentState.Phase[0].ChargingMAh = float64(register.value)
				case 1010:
					currentState.Phase[1].ChargingMAh = float64(register.value)
				case 1012:
					currentState.Phase[2].ChargingMAh = float64(register.value)
				case 1040:
					currentState.Phase[0].VoltageV = float64(register.value)
				case 1042:
					currentState.Phase[1].VoltageV = float64(register.value)
				case 1044:
					currentState.Phase[2].VoltageV = float64(register.value)
				}
			}
			registerFilled = true
			time.Sleep(60 * time.Second)
		}
	}()

	router := gin.Default()
	router.GET("/state", getState)
	err := router.Run(env.hostName + ":" + env.apiPort)
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msg("Get current state on http://" + env.hostName + ":" + env.apiPort + "/state")
	log.Fatal().Err(http.ListenAndServe(":"+env.apiPort, nil))

}

func initApp() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	initEnvironmentVariables()

	if env.debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug level activated.")
	}

	initRegisters()
}

func initEnvironmentVariables() {

	log.Info().Msg("Usage:")
	log.Info().Str("hostName", "Process listening on this hostName.").Msg("Mandatory environment parameter.")
	log.Info().Str("apiPort", "Process listening on this port.").Msg("Mandatory environment parameter.")
	log.Info().Str("wallboxName", "This is an IP or a servername.").Msg("Mandatory environment parameter.")
	log.Info().Str("wallboxPort", strconv.Itoa(modbusclient.MODBUS_PORT)).Msg("Optional: The port TCP/modbus listens.")

	log.Info().Str("debug", "false").Msg("Optional: Use debug mode for logging (true | false ). ")

	env.wallboxName = getEnv("wallboxName", "")
	if len(env.wallboxName) == 0 {
		log.Fatal().Msg("The environment variable wallboxName is unset. Please fix this.")
	}

	env.hostName = getEnv("hostName", "")
	if len(env.hostName) == 0 {
		log.Fatal().Msg("The environment variable hostName is unset. Please fix this.")
	}

	env.apiPort = getEnv("apiPort", "")
	if len(env.apiPort) == 0 {
		log.Fatal().Msg("The environment variable apiPort is unset. Please fix this.")
	}

	portString := getEnv("wallboxPort", strconv.Itoa(modbusclient.MODBUS_PORT))
	port, err := strconv.Atoi(portString)
	env.wallboxPort = port
	if err != nil {
		log.Fatal().Err(err).Str("wallboxPort", portString)
	}

	debug := getEnv("debug", "false")
	env.debug, err = strconv.ParseBool(debug)

	if err != nil {
		log.Fatal().Err(err).Str("debug", debug)
	}

	log.Info().Str("wallboxName", env.wallboxName).Msg("This is the configured wallboxName.")
	log.Info().Str("wallboxPort", strconv.Itoa(env.wallboxPort)).Msg("This is the configured port TCP/modbus listens.")
	log.Info().Str("debug", strconv.FormatBool(env.debug)).Msg("Log debug mode.")

}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func updateRegisterData() {

	log.Debug().Msg("updateRegisterData")

	conn, err := modbusclient.ConnectTCP(env.wallboxName, env.wallboxPort)
	if err != nil {
		log.Fatal().Err(err).Msg("Connection error.")
	}

	for i, register := range registers {
		log.Debug().Str("id", string(rune(registers[i].id))).Str("value", string(registers[i].value))
		registers[i] = readRegister(conn, register)
		log.Debug().Str("id", string(rune(registers[i].id))).Str("value", string(registers[i].value))
		time.Sleep(1 * time.Second)
	}

	modbusclient.DisconnectTCP(conn)

}

func readRegister(conn net.Conn, register register) register {

	// attempt to read one (0x01) holding registers starting at address 200
	readData := make([]byte, 3)
	readData[0] = byte(register.id >> 8)   // (High Byte)
	readData[1] = byte(register.id & 0xff) // (Low Byte)
	readData[2] = 0x01

	trace := zerolog.GlobalLevel() == zerolog.DebugLevel

	// make this read request transaction id 1, with a 300 millisecond tcp timeout
	readResult, readErr := modbusclient.TCPRead(conn, 300, 1, modbusclient.FUNCTION_READ_HOLDING_REGISTERS, false, 0x00, readData, trace)
	if readErr != nil {
		log.Fatal().Err(readErr)
	}

	var value int32
	var offset int

	from := len(readResult) - 4
	to := len(readResult)

	offset = 0
	value = 0

	for i := from; i < to; i++ {
		offset++
		switch offset {
		case 1:
			value = value + int32(readResult[i])*256*256*256
		case 2:
			value = value + int32(readResult[i])*256*256
		case 3:
			value = value + int32(readResult[i])*256
		case 4:
			value = value + int32(readResult[i])
		}
	}

	register.value = value

	return register
}
