package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"
)

type Movie struct {
	ID       string    `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type MovieStore struct {
	mu     sync.RWMutex
	movies []Movie
}

func NewMovieStore() *MovieStore {
	return &MovieStore{
		movies: []Movie{
			{
				ID:    "1",
				Isbn:  "345643",
				Title: "Toxic - A Fairy Tail For Grown-Ups",
				Director: &Director{
					Firstname: "Geetu Mohan",
					Lastname:  "Das",
				},
			},
			{
				ID:    "2",
				Isbn:  "452119",
				Title: "The Paradise",
				Director: &Director{
					Firstname: "Srikanth",
					Lastname:  "Odela",
				},
			},
		},
	}
}

func responsdJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	responsdJSON(w, status, map[string]string{"error": message})
}

func (s *MovieStore) getMovies(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	responsdJSON(w, http.StatusOK, s.movies)
}

func (s *MovieStore) getMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, movie := range s.movies {
		if movie.ID == id {
			responsdJSON(w, http.StatusOK, movie)
			return
		}
	}
	respondError(w, http.StatusNotFound, "Movie not Found")
}

func (s *MovieStore) updateMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updatedMovie Movie
	if err := json.NewDecoder(r.Body).Decode(&updatedMovie); err != nil {
		respondError(w, http.StatusBadGateway, "Inavalid JSON payload")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, movie := range s.movies {
		if movie.ID == id {
			updatedMovie.ID = id
			s.movies[index] = updatedMovie
			responsdJSON(w, http.StatusOK, updatedMovie)
			return
		}
	}
	respondError(w, http.StatusNotFound, "Movie Not Found")
}

func (s *MovieStore) createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	movie.ID = strconv.Itoa(rand.IntN(100))
	s.movies = append(s.movies, movie)

	responsdJSON(w, http.StatusCreated, movie)
}

func (s *MovieStore) deleteMovie(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, movie := range s.movies {
		if movie.ID == id {
			s.movies = slices.Delete(s.movies, index, index+1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	respondError(w, http.StatusNotFound, "Movie not Found")
}

func main() {
	store := NewMovieStore()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /movies", store.getMovies)
	mux.HandleFunc("GET /movies/{id}", store.getMovie)
	mux.HandleFunc("POST /movies", store.createMovie)
	mux.HandleFunc("PUT /movies/{id}", store.updateMovie)
	mux.HandleFunc("DELETE /movies/{id}", store.deleteMovie)

	server := &http.Server{
		Addr:         ":8000",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("Server Starting at PORT 8000")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
