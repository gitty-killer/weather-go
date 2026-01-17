package main

import (
  "bufio"
  "errors"
  "fmt"
  "os"
  "path/filepath"
  "strconv"
  "strings"
)

var Fields = []string{"day", "condition", "high", "low"}
var NumericField = "high"
var StorePath = filepath.Join("data", "store.txt")

func parseKV(items []string) (map[string]string, error) {
  record := map[string]string{}
  for _, item := range items {
    parts := strings.SplitN(item, "=", 2)
    if len(parts) != 2 {
      return nil, fmt.Errorf("invalid item: %s", item)
    }
    key, value := parts[0], parts[1]
    if !contains(Fields, key) {
      return nil, fmt.Errorf("unknown field: %s", key)
    }
    if strings.Contains(value, "|") {
      return nil, errors.New("value may not contain '|' ")
    }
    record[key] = value
  }
  for _, f := range Fields {
    if _, ok := record[f]; !ok {
      record[f] = ""
    }
  }
  return record, nil
}

func formatRecord(values map[string]string) string {
  parts := []string{}
  for _, k := range Fields {
    parts = append(parts, fmt.Sprintf("%s=%s", k, values[k]))
  }
  return strings.Join(parts, "|")
}

func parseLine(line string) (map[string]string, error) {
  values := map[string]string{}
  for _, part := range strings.Split(strings.TrimSpace(line), "|") {
    if part == "" {
      continue
    }
    kv := strings.SplitN(part, "=", 2)
    if len(kv) != 2 {
      return nil, fmt.Errorf("bad part: %s", part)
    }
    values[kv[0]] = kv[1]
  }
  return values, nil
}

func loadRecords() ([]map[string]string, error) {
  if _, err := os.Stat(StorePath); err != nil {
    return []map[string]string{}, nil
  }
  f, err := os.Open(StorePath)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  records := []map[string]string{}
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    line := strings.TrimSpace(scanner.Text())
    if line == "" {
      continue
    }
    r, err := parseLine(line)
    if err != nil {
      return nil, err
    }
    records = append(records, r)
  }
  return records, scanner.Err()
}

func appendRecord(values map[string]string) error {
  if err := os.MkdirAll(filepath.Dir(StorePath), 0755); err != nil {
    return err
  }
  f, err := os.OpenFile(StorePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
  if err != nil {
    return err
  }
  defer f.Close()
  _, err = f.WriteString(formatRecord(values) + "
")
  return err
}

func summary(records []map[string]string) string {
  count := len(records)
  if NumericField == "" {
    return fmt.Sprintf("count=%d", count)
  }
  total := 0
  for _, r := range records {
    n, err := strconv.Atoi(r[NumericField])
    if err == nil {
      total += n
    }
  }
  return fmt.Sprintf("count=%d, %s_total=%d", count, NumericField, total)
}

func contains(items []string, v string) bool {
  for _, item := range items {
    if item == v {
      return true
    }
  }
  return false
}

func main() {
  if len(os.Args) < 2 {
    fmt.Println("Usage: init | add key=value... | list | summary")
    os.Exit(2)
  }
  cmd := os.Args[1]
  args := os.Args[2:]
  switch cmd {
  case "init":
    _ = os.MkdirAll(filepath.Dir(StorePath), 0755)
    _ = os.WriteFile(StorePath, []byte(""), 0644)
  case "add":
    record, err := parseKV(args)
    if err != nil {
      fmt.Println(err)
      os.Exit(2)
    }
    if err := appendRecord(record); err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  case "list":
    records, err := loadRecords()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    for _, r := range records {
      fmt.Println(formatRecord(r))
    }
  case "summary":
    records, err := loadRecords()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
    fmt.Println(summary(records))
  default:
    fmt.Println("Unknown command:", cmd)
    os.Exit(2)
  }
}
