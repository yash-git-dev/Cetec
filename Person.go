package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Person struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	City        string `json:"city"`
	State       string `json:"state"`
	Street1     string `json:"street1"`
	Street2     string `json:"street2"`
	ZipCode     string `json:"zip_code"`
}

func PersonGET(c *gin.Context) {
	var person Person

	personID := c.Param("person_id")
	if personID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Person ID is required"})
	}

	db, err := ConnectDB()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong" + err.Error()})
	}
	defer db.Close()

	query, err := db.Prepare(`SELECT p.name, ph.number, addr.city, addr.state, addr.street1, addr.street2, addr.zip_code 
		FROM person p JOIN phone ph on p.id = ph.person_id JOIN address_join adj on adj.person_id = p.id 
		JOIN address addr on addr.id = adj.address_id WHERE p.id = ?`)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	}
	defer query.Close()

	err = query.QueryRow(personID).Scan(
		&person.Name, &person.PhoneNumber, &person.City, &person.State, &person.Street1, &person.Street2, &person.ZipCode,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	}

	c.JSON(http.StatusOK, person)
}

func PersonPOST(c *gin.Context) {
	var person Person

	err := c.ShouldBindJSON(&person)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "please check input"})
	}

	db, err := ConnectDB()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	}

	// Person insertion
	_, err = tx.Exec("INSERT INTO person (name) VALUES (?)", person.Name)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	var personID int
	err = tx.QueryRow("SELECT id from person where name = ? order by id DESC limit 1", person.Name).Scan(&personID)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	// Phone insertion
	_, err = tx.Exec("INSERT INTO phone (person_id, number) VALUES (?, ?)", personID, person.PhoneNumber)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	//Address insertion
	_, err = tx.Exec("INSERT INTO address (city, state, street1, street2, zip_code) VALUES (?, ?, ?, ?, ?)",
		person.City, person.State, person.Street1, person.Street2, person.ZipCode)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	var addressID int
	err = tx.QueryRow("SELECT id from address where street1 = ? and street2 = ? order by id DESC limit 1",
		person.Street1, person.Street2).Scan(&addressID)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	//Address Join insertion
	_, err = tx.Exec("INSERT INTO address_join (person_id, address_id) VALUES (?, ?)", personID, addressID)
	if err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	if err := tx.Commit(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error while inserting data"})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data inserted successfully"})
}
