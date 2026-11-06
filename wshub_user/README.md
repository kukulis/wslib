# running tests in
    
    docker compose exec -u 1000:1000 wshub_user bash -c "cd wshub_user && vendor/bin/phpunit tests"