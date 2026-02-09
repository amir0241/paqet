package client

import (
	"testing"
	"paqet/internal/conf"
	"paqet/internal/pkg/iterator"
)

func TestNewConn_WithNoConnections(t *testing.T) {
	// Create a client with an empty iterator
	c := &Client{
		cfg:  &conf.Conf{},
		iter: &iterator.Iterator[*timedConn]{},
	}
	
	// Attempt to create a new connection
	conn, err := c.newConn()
	
	// Should return error and not panic
	if err == nil {
		t.Error("Expected error when no connections available, got nil")
	}
	if conn != nil {
		t.Errorf("Expected nil connection when error occurs, got %v", conn)
	}
	
	// Check error message
	expectedErrMsg := "no available connections"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
