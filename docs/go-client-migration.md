# Mechanical Migration Guide

This guide demonstrates how to convert existing Go applications to use ConsoleHub with minimal code changes.

---

## 1. Standard Output Migration (`fmt` -> `consolehub`)

### Before
```go
package main

import "fmt"

func main() {
    fmt.Println("Starting process...")
    fmt.Printf("Processing item %d of %d\n", 5, 100)
}
```

### After
```go
package main

import "consolehub/libraries/go/consolehub"

func main() {
    defer consolehub.Close()

    consolehub.Println("Starting process...")
    consolehub.Printf("Processing item %d of %d\n", 5, 100)
}
```

---

## 2. Standard Logger Migration (`log` -> `consolehub`)

### Before
```go
package main

import "log"

func main() {
    log.Println("Starting server...")
}
```

### After
```go
package main

import "consolehub/libraries/go/consolehub"

func main() {
    defer consolehub.Close()

    consolehub.Infof("Starting server...")
}
```

---

## 3. Writer Wrapping (`io.Writer`)

### Before
```go
mw := io.MultiWriter(os.Stdout, file)
```

### After
```go
mw := io.MultiWriter(os.Stdout, consolehub.Stdout(), file)
```

---

## 4. What NOT to Change

Do **NOT** replace standard Go error handling or formatting functions:
- `fmt.Sprintf` -> keep as `fmt.Sprintf`
- `fmt.Errorf` -> keep as `fmt.Errorf`
- `errors.Is` -> keep as `errors.Is`
- `errors.As` -> keep as `errors.As`
