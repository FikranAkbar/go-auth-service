package logger

// Example usage of the logger package
//
// Basic usage:
//   logger.Info("Server starting")
//   logger.Errorf("Failed to connect: %v", err)
//
// With configuration:
//   log := logger.New(logger.Config{
//       Level:       "debug",
//       Environment: "production",
//       ServiceName: "go-auth-service",
//       Version:     "1.0.0",
//   })
//   log.Info("Application initialized")
//
// With context:
//   log := logger.WithContext(ctx).WithField("user_id", userID)
//   log.Info("User logged in")
//
// With fields:
//   logger.WithFields(map[string]interface{}{
//       "user_id": "123",
//       "action": "login",
//   }).Info("Authentication successful")
//
// HTTP logging:
//   log.HTTP("GET", "/api/users", 200, time.Since(start), clientIP)
//
// Database logging:
//   log.Database("SELECT", "SELECT * FROM users WHERE id = ?", duration, err)
//
// Auth logging:
//   log.Auth(userID, "login", true, "")
//
// Performance logging:
//   log.Performance("cache_lookup", duration, map[string]interface{}{
//       "cache_hit": true,
//       "key": "user:123",
//   })
