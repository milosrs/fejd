package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "ok")
    })

    fmt.Println("backend listening on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        panic(err)
    }
}
