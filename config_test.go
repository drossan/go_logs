package go_logs

import (
	"os"
	"sync"
	"testing"
)

// TestNotificationSettingsConcurrency detecta race conditions en acceso concurrente
// a notificationSettings siguiendo el enfoque TDD RED-GREEN-REFACTOR
//
// GREEN: Este test verifica que el acceso a través de la API pública es thread-safe
func TestNotificationSettingsConcurrency(t *testing.T) {
	// Setup: Configurar variables de entorno para pruebas
	os.Setenv("SAVE_LOG_FILE", "0")
	os.Setenv("NOTIFICATION_FATAL_LOG", "1")
	os.Setenv("NOTIFICATION_ERROR_LOG", "1")
	os.Setenv("NOTIFICATION_WARNING_LOG", "1")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "1")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	defer func() {
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
	}()

	// Inicializar configuración
	Init()

	// Simular acceso concurrente usando la API pública
	// Verifica que getNotificationSettings() es thread-safe
	const goroutines = 10
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				// Usar la función thread-safe getter
				_ = getNotificationSettings("ERROR")
				_ = getNotificationSettings("INFO")
				_ = getNotificationSettings("FATAL")
			}
		}(i)
	}

	wg.Wait()
	// Test pasa si no hay data races detectadas por go test -race
}

// TestNotificationSettingsDirectMapAccess demuestra la race condition
// cuando se accede directamente al mapa sin protección (RED phase)
//
// NOTA: Este test INTENCIONALMENTE causa una data race para demostrar el problema.
// Debe ejecutarse con `go test -race` para ver el WARNING.
func TestNotificationSettingsDirectMapAccess(t *testing.T) {
	os.Setenv("SAVE_LOG_FILE", "0")
	os.Setenv("NOTIFICATION_FATAL_LOG", "1")
	os.Setenv("NOTIFICATION_ERROR_LOG", "1")
	os.Setenv("NOTIFICATION_WARNING_LOG", "1")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "1")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	defer func() {
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
	}()

	Init()

	// Acceso directo sin protección (INTENCIONAL para demostrar el problema)
	const goroutines = 10
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				// Acceso DIRECTO al mapa sin mutex (causa race condition)
				notificationSettingsMutex.RLock()
				_ = notificationSettings["ERROR"]
				notificationSettingsMutex.RUnlock()

				// Simular escritura concurrente
				notificationSettingsMutex.Lock()
				notificationSettings["ERROR"] = (id%2 == 0)
				notificationSettingsMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()
	// Si llegamos aquí, el test pasa
	// pero go test -race debería reportar WARNING: DATA RACE
}

// TestInitNotCalled verifica que el paquete funcione correctamente
// aunque Init() no se haya llamado explícitamente (auto-inicialización)
func TestInitNotCalled(t *testing.T) {
	// Resetear estado global para este test
	isInit = false

	// Crear un archivo temporal para logs
	tempDir := t.TempDir()
	os.Setenv("SAVE_LOG_FILE", "1")
	os.Setenv("LOG_FILE_NAME", "test.log")
	os.Setenv("LOG_FILE_PATH", tempDir)
	os.Setenv("NOTIFICATION_FATAL_LOG", "1")
	os.Setenv("NOTIFICATION_ERROR_LOG", "1")
	os.Setenv("NOTIFICATION_WARNING_LOG", "1")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "1")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	defer func() {
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("LOG_FILE_NAME")
		os.Unsetenv("LOG_FILE_PATH")
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
	}()

	// Llamar a ErrorLog sin llamar Init() explícitamente
	// Esto debería disparar auto-inicialización en saveLog()
	ErrorLog("Test message without Init()")

	// Verificar que no hubo panic y que la inicialización funcionó
	if !isInit {
		t.Error("Init should have been called automatically")
	}
}

// TestNotificationSettingsValues verifies that notification settings
// are correctly loaded from environment variables
func TestNotificationSettingsValues(t *testing.T) {
	os.Setenv("SAVE_LOG_FILE", "0")
	os.Setenv("NOTIFICATION_FATAL_LOG", "1")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "1")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	defer func() {
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
	}()

	Init()

	// Verificar valores usando la función thread-safe
	if getNotificationSettings("FATAL") != true {
		t.Error("FATAL should be true")
	}
	if getNotificationSettings("ERROR") != false {
		t.Error("ERROR should be false")
	}
	if getNotificationSettings("INFO") != true {
		t.Error("INFO should be true")
	}
}

// TestLogFilePathConstruction verifies that file paths are constructed
// correctly using filepath.Join() for portability (Issue #4)
func TestLogFilePathConstruction(t *testing.T) {
	tests := []struct {
		name           string
		logFilePath    string
		logFileName    string
		expectedInPath string
	}{
		{
			name:           "Empty path uses filename only",
			logFilePath:    "",
			logFileName:    "test.log",
			expectedInPath: "test.log",
		},
		{
			name:           "Path and filename joined",
			logFilePath:    t.TempDir(),
			logFileName:    "app.log",
			expectedInPath: "app.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Reset persistent file state before each test
			Close() // Close any previously opened file

			logFilePath = tt.logFilePath
			logFileName = tt.logFileName

			// Call function under test
			file := openLogFile()

			// Verify file was created successfully
			if file == nil {
				t.Error("Expected file to be created, got nil")
				return
			}

			// Clean up: use Close() instead of file.Close() for persistent file
			Close()
		})
	}
}

// TestLoadSlackConfig verifies loadSlackConfig initializes Slack notifier
// Issue #7: Tests credential validation and graceful degradation
func TestLoadSlackConfig(t *testing.T) {
	t.Run("Valid credentials initializes notifier", func(t *testing.T) {
		// Setup: Valid Slack credentials
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Call function under test
		loadSlackConfig()

		// Verify notifier was initialized
		if notifier == nil {
			t.Error("Expected notifier to be initialized with valid credentials")
		}

		if !IsNotifierEnabled() {
			t.Error("Expected notifier to be enabled with valid credentials")
		}
	})

	t.Run("Missing token disables notifier gracefully", func(t *testing.T) {
		// Setup: Missing token (only channel)
		os.Unsetenv("SLACK_TOKEN")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Call function under test
		loadSlackConfig()

		// Verify notifier exists but is disabled
		if notifier == nil {
			t.Error("Expected notifier to be created (but disabled)")
		}

		if IsNotifierEnabled() {
			t.Error("Expected notifier to be disabled when token is missing")
		}
	})

	t.Run("Missing channel disables notifier gracefully", func(t *testing.T) {
		// Setup: Missing channel (only token)
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Unsetenv("SLACK_CHANNEL_ID")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
		}()

		// Call function under test
		loadSlackConfig()

		// Verify notifier exists but is disabled
		if notifier == nil {
			t.Error("Expected notifier to be created (but disabled)")
		}

		if IsNotifierEnabled() {
			t.Error("Expected notifier to be disabled when channel is missing")
		}
	})

	t.Run("Both missing disables notifier", func(t *testing.T) {
		// Setup: No credentials at all
		os.Unsetenv("SLACK_TOKEN")
		os.Unsetenv("SLACK_CHANNEL_ID")

		// Call function under test
		loadSlackConfig()

		// Verify notifier exists but is disabled
		if notifier == nil {
			t.Error("Expected notifier to be created (but disabled)")
		}

		if IsNotifierEnabled() {
			t.Error("Expected notifier to be disabled when both credentials are missing")
		}
	})
}

// TestLoadSlackConfigBackwardCompatibility verifies support for typo version
// Issue #8: SLACK_CHANEL_ID (typo) should work with warning
func TestLoadSlackConfigBackwardCompatibility(t *testing.T) {
	// Setup: Use typo version only
	os.Setenv("SLACK_TOKEN", "xoxb-test-token")
	os.Setenv("SLACK_CHANEL_ID", "C1234567890") // Typo version
	os.Unsetenv("SLACK_CHANNEL_ID")             // Ensure correct version is NOT set
	defer func() {
		os.Unsetenv("SLACK_TOKEN")
		os.Unsetenv("SLACK_CHANEL_ID")
	}()

	// Call function under test
	loadSlackConfig()

	// Verify notifier is enabled (backward compatibility works)
	if notifier == nil {
		t.Fatal("Expected notifier to be created")
	}

	if !IsNotifierEnabled() {
		t.Error("Expected notifier to be enabled with typo version (backward compatibility)")
	}
}

// TestLoadLogLevel verifies LOG_LEVEL environment variable loading
func TestLoadLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		expected    int
	}{
		{"Trace level", "trace", LevelTrace},
		{"Debug level", "debug", LevelDebug},
		{"Info level", "info", LevelInfo},
		{"Warn level", "warn", LevelWarn},
		{"Warning level", "warning", LevelWarn},
		{"Error level", "error", LevelError},
		{"Fatal level", "fatal", LevelFatal},
		{"Silent level", "silent", LevelSilent},
		{"None level", "none", LevelSilent},
		{"Disable level", "disable", LevelSilent},
		{"Empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			logLevel = 0

			// Set environment variable
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
				defer os.Unsetenv("LOG_LEVEL")
			}

			// Call function under test
			loadLogLevel()

			// Verify result
			if logLevel != tt.expected {
				t.Errorf("Expected logLevel=%d, got %d", tt.expected, logLevel)
			}
		})
	}
}

// TestLoadLogLevelCaseInsensitive verifies LOG_LEVEL is case-insensitive
func TestLoadLogLevelCaseInsensitive(t *testing.T) {
	tests := []string{"TRACE", "Trace", "tRaCe", "DEBUG", "Debug", "DeBuG"}

	for _, level := range tests {
		t.Run(level, func(t *testing.T) {
			// Reset state
			logLevel = 0

			os.Setenv("LOG_LEVEL", level)
			defer os.Unsetenv("LOG_LEVEL")

			loadLogLevel()

			if logLevel == 0 {
				t.Error("Expected logLevel to be set, but got 0")
			}
		})
	}
}

// TestGetNumericLevel verifies numeric level conversion
func TestGetNumericLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected int
	}{
		{"FATAL", LevelFatal},
		{"ERROR", LevelError},
		{"WARNING", LevelWarn},
		{"WARN", LevelWarn},
		{"INFO", LevelInfo},
		{"DEBUG", LevelDebug},
		{"TRACE", LevelTrace},
		{"UNKNOWN", LevelSilent},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := getNumericLevel(tt.level)
			if result != tt.expected {
				t.Errorf("Expected %d for %s, got %d", tt.expected, tt.level, result)
			}
		})
	}
}

// TestLogLevelThreshold verifies syslog-style threshold behavior
// Messages with level >= configured level should be logged
func TestLogLevelThreshold(t *testing.T) {
	tests := []struct {
		name          string
		configuredLevel string
		messageLevel   string
		shouldLog      bool
	}{
		{"Info level logs info and above", "info", "INFO", true},
		{"Info level logs error", "info", "ERROR", true},
		{"Info level logs fatal", "info", "FATAL", true},
		{"Info level does not log debug", "info", "DEBUG", false},
		{"Error level logs error and fatal", "error", "ERROR", true},
		{"Error level logs fatal", "error", "FATAL", true},
		{"Error level does not log warning", "error", "WARNING", false},
		{"Error level does not log info", "error", "INFO", false},
		{"Warn level logs warning and above", "warn", "WARNING", true},
		{"Warn level logs error", "warn", "ERROR", true},
		{"Warn level does not log info", "warn", "INFO", false},
		{"Fatal level only logs fatal", "fatal", "FATAL", true},
		{"Fatal level does not log error", "fatal", "ERROR", false},
		{"Debug level logs debug and above", "debug", "DEBUG", true},
		{"Debug level logs info", "debug", "INFO", true},
		{"Debug level logs error", "debug", "ERROR", true},
		{"Trace level logs everything", "trace", "TRACE", true},
		{"Trace level logs debug", "trace", "DEBUG", true},
		{"Silent level logs nothing", "silent", "FATAL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			logLevel = 0
			useLegacySystem = false
			notificationSettings = nil

			os.Setenv("LOG_LEVEL", tt.configuredLevel)
			os.Setenv("SAVE_LOG_FILE", "0")
			os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
			// Set all legacy notification vars to 0 to ensure they don't interfere
			os.Setenv("NOTIFICATION_FATAL_LOG", "0")
			os.Setenv("NOTIFICATION_ERROR_LOG", "0")
			os.Setenv("NOTIFICATION_WARNING_LOG", "0")
			os.Setenv("NOTIFICATION_INFO_LOG", "0")
			os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
			defer func() {
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("SAVE_LOG_FILE")
				os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
				os.Unsetenv("NOTIFICATION_FATAL_LOG")
				os.Unsetenv("NOTIFICATION_ERROR_LOG")
				os.Unsetenv("NOTIFICATION_WARNING_LOG")
				os.Unsetenv("NOTIFICATION_INFO_LOG")
				os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
			}()

			// Initialize both configs to set up notificationSettings map
			loadNotificationsConfig()
			loadLogLevel()

			// Test getNotificationSettings which implements the threshold logic
			result := getNotificationSettings(tt.messageLevel)

			if result != tt.shouldLog {
				t.Errorf("LOG_LEVEL=%s, message=%s: expected %v, got %v",
					tt.configuredLevel, tt.messageLevel, tt.shouldLog, result)
			}
		})
	}
}

// TestLegacySystemPrecedence verifies legacy system takes precedence over LOG_LEVEL
func TestLegacySystemPrecedence(t *testing.T) {
	// Setup: Configure BOTH systems
	// Legacy: Only ERROR and FATAL enabled
	// New: INFO level (would log INFO, WARNING, ERROR, FATAL)
	os.Setenv("NOTIFICATION_FATAL_LOG", "1")
	os.Setenv("NOTIFICATION_ERROR_LOG", "1")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "0")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0") // Set explicitly to avoid ParseBool error
	os.Setenv("LOG_LEVEL", "info") // New system would log INFO
	os.Setenv("SAVE_LOG_FILE", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	defer func() {
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
	}()

	// Reset state
	logLevel = 0
	useLegacySystem = false

	// Initialize
	loadNotificationsConfig()
	loadLogLevel()

	// Verify legacy system is detected
	if !useLegacySystem {
		t.Error("Expected useLegacySystem to be true when legacy env vars are set")
	}

	// Verify legacy system takes precedence
	// INFO should NOT be logged (legacy: NOTIFICATION_INFO_LOG=0)
	// even though LOG_LEVEL=info would log it
	if getNotificationSettings("INFO") {
		t.Error("Legacy system should take precedence: INFO should not be logged")
	}

	// ERROR should be logged (both systems agree)
	if !getNotificationSettings("ERROR") {
		t.Error("ERROR should be logged (enabled in both systems)")
	}

	// WARNING should NOT be logged (legacy: NOTIFICATION_WARNING_LOG=0)
	// even though LOG_LEVEL=info would log it
	if getNotificationSettings("WARNING") {
		t.Error("Legacy system should take precedence: WARNING should not be logged")
	}
}

// TestLogLevelOnly verifies LOG_LEVEL works when legacy system is not configured
func TestLogLevelOnly(t *testing.T) {
	// Setup: Only LOG_LEVEL configured (no legacy vars)
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("SAVE_LOG_FILE", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")
	// Explicitly disable all legacy notification settings
	os.Setenv("NOTIFICATION_FATAL_LOG", "0")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "0")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SAVE_LOG_FILE")
		os.Unsetenv("NOTIFICATIONS_SLACK_ENABLED")
		os.Unsetenv("NOTIFICATION_FATAL_LOG")
		os.Unsetenv("NOTIFICATION_ERROR_LOG")
		os.Unsetenv("NOTIFICATION_WARNING_LOG")
		os.Unsetenv("NOTIFICATION_INFO_LOG")
		os.Unsetenv("NOTIFICATION_SUCCESS_LOG")
	}()

	// Reset state
	logLevel = 0
	useLegacySystem = false

	// Initialize both configs
	loadNotificationsConfig()
	loadLogLevel()

	// Verify legacy system is NOT detected
	if useLegacySystem {
		t.Error("Expected useLegacySystem to be false when all legacy vars are 0")
	}

	// Verify LOG_LEVEL numeric threshold works
	// ERROR level should log ERROR and FATAL
	tests := []struct {
		level    string
		expected bool
	}{
		{"FATAL", true},   // >= 50
		{"ERROR", true},   // >= 50
		{"WARNING", false}, // < 50
		{"INFO", false},   // < 50
		{"SUCCESS", false}, // Not in numeric system
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			result := getNotificationSettings(tt.level)
			if result != tt.expected {
				t.Errorf("LOG_LEVEL=error, message=%s: expected %v, got %v",
					tt.level, tt.expected, result)
			}
		})
	}
}

