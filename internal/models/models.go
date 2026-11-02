package models

// All lists every model managed by AutoMigrate, in dependency order.
func All() []interface{} {
	return []interface{}{
		&User{},
		&Event{},
	}
}
