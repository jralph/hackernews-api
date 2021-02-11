# Hackernews Data Scraper + API

Run using the following command:

```bash
docker-compose up -d --remove-orphans --build && docker-compose ps
```

The end output of this will be something like the following:

```text
              Name                            Command                  State                Ports
----------------------------------------------------------------------------------------------------------
hackernews-api_api_1               /api -redis-host=redis:6379      Up             0.0.0.0:55032->8901/tcp
hackernews-api_redis-commander_1   /usr/bin/dumb-init -- /red ...   Up (healthy)   0.0.0.0:55021->8081/tcp
hackernews-api_redis_1             docker-entrypoint.sh redis ...   Up             0.0.0.0:55019->6379/tcp
hackernews-api_scraper_1           /scraper -redis-host=redis ...   Up             8901/tcp
```

You can browse to the api at `0.0.0.0:55032` for example.

You can also access a UI for redis by browsing to `0.0.0.0:55021` for example.