# Algebra APR Backend

This is a backend service for calculating and serving APR (Annual Percentage Rate) data for Algebra pools and eternal farmings across different blockchain networks.

## Quick Start

### First Time Setup

1. **SSL Certificates**: Place your SSL certificates in the `nginx/certs/` directory before starting the application.

2. **Network Configuration**: Add your blockchain network configuration to `config.json`:
   ```json
   {
     "port": "8080",
     "log_level": "info", 
     "apr_update_minutes": 30,
     "networks": [
       {
         "title": "YourNetworkName",
         "analytics_subgraph_url": "https://your-analytics-subgraph-url",
         "subgraph_farming_url": "https://your-farming-subgraph-url",
         "api_key": "your-api-key-if-required"
       }
     ]
   }
   ```

3. **Initial Setup**: Run the following command to set up the database and start the application:
   ```bash
   make migrate-and-run
   ```
   This command will:
   - Stop any running containers
   - Rebuild all containers
   - Start the database
   - Run database migrations
   - Start the application and nginx

### Development

To restart just the application container (useful during development):
```bash
make rebuild c=app
```

## API Endpoints

The API provides the following endpoints for retrieving APR and TVL data:

### Pools

- **GET** `/api/pools/apr?network=<network-title>`
  - Returns the current APR for all pools in the specified network
  - Response format: `{"pool_address": apr_value, ...}`

- **GET** `/api/pools/max-apr?network=<network-title>`
  - Returns the maximum APR for all pools in the specified network
  - Response format: `{"pool_address": max_apr_value, ...}`

### Eternal Farmings

- **GET** `/api/eternal-farmings/apr?network=<network-title>`
  - Returns the current APR for all eternal farmings in the specified network
  - Response format: `{"farming_hash": apr_value, ...}`

- **GET** `/api/eternal-farmings/max-apr?network=<network-title>`
  - Returns the maximum APR for all eternal farmings in the specified network
  - Response format: `{"farming_hash": max_apr_value, ...}`

- **GET** `/api/eternal-farmings/tvl?network=<network-title>`
  - Returns the Total Value Locked (TVL) for all eternal farmings in the specified network
  - Response format: `{"farming_hash": tvl_value, ...}`

### Parameters

- `network` (query parameter): The blockchain network name (e.g., "Polygon", "Berachain")

# CORS enabled
# CORS enabled
