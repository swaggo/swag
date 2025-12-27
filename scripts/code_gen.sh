#!/bin/bash

# Get the package name from go.mod by extracting the module line and removing 'github.com/'
PACKAGE_NAME=$(grep '^module ' go.mod | sed 's/module github.com\///')

# Echo the package name to make sure it's correct
echo "Package detected: $PACKAGE_NAME"

# Function to create an object
create_object() {
  echo "Enter the singular model name (Lowercase, snake case, non-plural):"
  read model_name
  echo "Enter the plural form of the model name (Lowercase, snake case):"
  read model_plural

  core_gen object "${model_name}" "-plural=${model_plural}" "-modelPackage=${PACKAGE_NAME}"

  go generate "./internal/models/${model_name}/${model_name}.go"
  go generate "./internal/controllers/${model_plural}/setup.go"
}

# Function to create a public object
create_public_object() {
  echo "Enter the singular model name (Lowercase, snake case, non-plural):"
  read model_name
  echo "Enter the plural form of the model name (Lowercase, snake case):"
  read model_plural

  core_gen object "${model_name}" "-plural=${model_plural}" "-modelPackage=${PACKAGE_NAME}" "-public=true"

  go generate "./internal/models/${model_name}/${model_name}.go"
  go generate "./internal/controllers/${model_plural}/setup.go"
}

# Function to create a migration
create_migration() {
  echo "What's the model name or "seed" for global data? (Lowercase, snake case, non-plural):"
  read model_name
  echo "Enter a descriptive name for the migration (Lowercase, snake case):"
  read migration_name

  # Since there's no corresponding command in the core_gen main.go for migrations,
  # you might need to implement this functionality or adjust as needed
  core_gen migration "${model_name}" "-label=${migration_name}"
}

# Main menu
echo "What do you want to create for ${PACKAGE_NAME}?"
echo "1) Create Internal Object"
echo "2) Create Public Object"
echo "3) Create Migration"
read choice

case $choice in
  1)
    create_object
    ;;
  2)
    create_public_object
    ;;
  3)
    create_migration
    ;;
  *)
    echo "Invalid option. Please select 1, 2, or 3."
    ;;
esac