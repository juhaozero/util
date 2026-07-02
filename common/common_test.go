package common

import (
	"fmt"
	"testing"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestArrayToString(t *testing.T) {
	array := []any{1, 2, 3, 4, 5}
	str := ArrayToString(array)
	if str != "1,2,3,4,5" {
		t.Fatalf("ArrayToString = %s, want %s", str, "1,2,3,4,5")
	}
	fmt.Printf("ArrayToString = %s\n", str)
	t.Logf("ArrayToString = %s", str)
}

func TestStructToMapString(t *testing.T) {
	user := User{Name: "John", Age: 20}
	m, err := StructToMapString(user)
	if err != nil {
		t.Fatalf("StructToMapString error: %v", err)
	}
	if m["name"] != "John" || m["age"] != "20" {
		t.Fatalf("StructToMapString = %#v", m)
	}

	var empty User
	if _, err := StructToMapString(&empty); err != nil {
		t.Fatalf("StructToMapString empty struct error: %v", err)
	}

	var nilUser *User
	if _, err := StructToMapString(nilUser); err == nil {
		t.Fatal("StructToMapString(nil) should return error")
	}
}

func TestMapToStruct(t *testing.T) {
	var user User

	err := MapToStruct(map[string]string{"name": "John", "age": "20"}, &user)
	if err != nil {
		t.Fatalf("MapToStruct error: %v", err)
	}
	fmt.Println("user", user)

	if user.Name != "John" {
		t.Fatalf("MapToStruct error: user.Name = %s, want %s", user.Name, "John")
	}
	if user.Age != 20 {
		t.Fatalf("MapToStruct error: user.Age = %d, want %d", user.Age, 20)
	}
}
