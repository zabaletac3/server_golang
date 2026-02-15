#!/bin/bash

# Simplifica los nombres de modelos en swagger.json
SWAGGER_FILE="internal/app/docs/swagger.json"

if [ -f "$SWAGGER_FILE" ]; then
    sed -i \
        -e 's/internal_modules_users\.//g' \
        -e 's/internal_modules_auth\.//g' \
        -e 's/internal_modules_permissions\.//g' \
        -e 's/internal_modules_resources\.//g' \
        -e 's/internal_modules_roles\.//g' \
        -e 's/internal_modules_tenant\.//g' \
        -e 's/internal_modules_plans\.//g' \
        -e 's/internal_modules_payments\.//g' \
        -e 's/github_com_eren_dev_go_server_internal_shared_validation\.//g' \
        -e 's/github_com_eren_dev_go_server_internal_shared_pagination\.//g' \
        "$SWAGGER_FILE"
    
    echo "✅ swagger.json simplified"
fi
