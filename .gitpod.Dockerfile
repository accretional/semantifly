FROM gitpod/workspace-full

USER root
RUN apt-get update && \
    apt-get install -y postgresql postgresql-contrib

RUN service postgresql start && \
    sudo -u postgres psql -c "CREATE USER gitpod WITH PASSWORD 'gitpod';" && \
    sudo -u postgres psql -c "CREATE DATABASE semantifly;" && \
    sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE semantifly TO gitpod;"

USER gitpod

EXPOSE 5432

CMD ["sudo", "service", "postgresql", "start"]
