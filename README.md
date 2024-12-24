# RSS reader

This is the backend for my personal RSS reader project.

It uses `gofeed` library under the hood to parse the feeds and sqlite as the database.

For the frontend, check out the Android application written in Flutter

## Structure

### /server

The actual route handler functions

- /server/middleware.go: middleware for handling authentication

### /token

Token structure with DB logic

### /refresh

A task that fetches new articles from the feeds and adds them to the database, runs with a specified periodicity

### /migrations

Database migrations

### /feed

Feed and article structures along with DB logic and utilities for converting from `gofeed` types

### /util

Miscellaneous functions used randomly in the project
