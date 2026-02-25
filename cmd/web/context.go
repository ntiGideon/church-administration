package main

type contextKey string

const isAuthenticatedKey = contextKey("isAuthenticated")
const userContextKey = contextKey("user")
