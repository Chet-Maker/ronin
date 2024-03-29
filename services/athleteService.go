package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ronin/models"

	"strconv"

	"ronin/repositories"

	"github.com/gorilla/mux"
)

var athleteRepo *repositories.AthleteRepository

func SetAthleteRepo(r *repositories.AthleteRepository) {
	athleteRepo = r
}

type AthleteUsername struct {
	Username string `json:"username" db:"username"`
}

type AthleteId struct {
	AthleteId int `json:"athleteId" db:"athlete_id"`
}

func GetAllAthleteUsernames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	usernames, err := athleteRepo.GetAllUsernames()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(&usernames)
}

func GetAllAthletes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	athletes, err := athleteRepo.GetAllAthletes()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(&athletes)
}

func GetAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["athlete_id"]
	athletes, err := athleteRepo.GetAthleteById(id)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(&athletes)
}

func GetAthleteByUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	username := vars["username"]
	athlete, err := athleteRepo.GetAthleteByUsername(username)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 400)
		return
	}

	json.NewEncoder(w).Encode(&athlete)
}

func IsAuthorizedUser(w http.ResponseWriter, r *http.Request) {
	var athlete models.Athlete
	err := json.NewDecoder(r.Body).Decode(&athlete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	isAuthorized, returnedAthlete, err := athleteRepo.IsAuthorizedUser(athlete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if !isAuthorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	} else if isAuthorized {
		idObj := AthleteId{AthleteId: returnedAthlete.AthleteId}
		json.NewEncoder(w).Encode(&idObj)
	} else {
		json.NewEncoder(w).Encode(false)
	}
}

func CreateAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var athlete models.Athlete
	err := json.NewDecoder(r.Body).Decode(&athlete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	athleteId, err := athleteRepo.CreateAthlete(athlete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&athleteId)
}

func UpdateAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var athlete models.Athlete
	err := json.NewDecoder(r.Body).Decode(&athlete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `UPDATE athlete SET first_name = $2, last_name = $3, username = $4, birth_date = $5, email = $6, password = $7 WHERE athlete_id = $8`
	_, err = dbconn.Queryx(sqlStatement, athlete.FirstName, athlete.LastName, athlete.Username, athlete.BirthDate, athlete.Email, athlete.Password, athlete.AthleteId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Athlete updated successfully")

}

func DeleteAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	id := vars["athlete_id"]

	sqlStatement := `DELETE FROM athlete WHERE athlete_id = $1`
	_, err := dbconn.Queryx(sqlStatement, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Athlete deleted successfully")

}

func GetAthleteRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var record models.AthleteRecord
	vars := mux.Vars(r)
	id := vars["athlete_id"]
	sqlStmt := `SELECT * FROM athlete_record where athlete_id = $1`
	row, err := dbconn.Queryx(sqlStmt, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		for row.Next() {
			err2 := row.StructScan(&record)
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		json.NewEncoder(w).Encode(&record)
	}
}

func FollowAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var follow models.Follow
	err := json.NewDecoder(r.Body).Decode(&follow)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sqlStatement := `INSERT INTO following (follower_id, followed_id) VALUES ($1, $2)`
	_, err = dbconn.Queryx(sqlStatement, follow.FollowerId, follow.FollowedId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Athlete followed successfully")
}

func UnfollowAthlete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	followerId, err := strconv.Atoi(vars["followerId"])
	if err != nil {
		http.Error(w, "Invalid followerId", http.StatusBadRequest)
		return
	}
	followedId, err := strconv.Atoi(vars["followedId"])
	if err != nil {
		http.Error(w, "Invalid followedId", http.StatusBadRequest)
		return
	}
	_, err = repo.UnfollowAthlete(followerId, followedId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetAthletesFollowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var follows []int
	vars := mux.Vars(r)
	id := vars["athlete_id"]
	var tempFollow = models.GetFollow()
	sqlStmt := `SELECT * FROM following where follower_id = $1`
	rows, err := dbconn.Queryx(sqlStmt, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		err2 := rows.StructScan(&tempFollow)
		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusBadRequest)
			return
		}
		follows = append(follows, tempFollow.FollowedId)
	}

	json.NewEncoder(w).Encode(follows)
}
