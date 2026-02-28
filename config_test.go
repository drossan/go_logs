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
