# Semantifly

Semantifly is a command line tool for implementing local RAG for the GenAI-for-coding use case. 

Users add data, or entire data sources like code repos and cloud projects, and Semantifly automatically gathers all the data corresponding to that data source (such as databases and other resources within a cloud project), serializes them as text and vector embeddings, and adds them to a local vector database. It then provides the necessary funcitonality for this database to be integrated with a LLM to implement Retrieval Augmented Generation with the index data - instead of having to manually tell the LLM what your database schema is, or the structure of an API in another repo within your organization, it can automatically retrive and access that data.

This is currently a work in progress.

# Intended Usage

Users should be able to operate Semantifly similar to the following example CLI invocations:

## Basic CRUD

Add data: ```semantifly add data repo/README.md --type internal_documentation```

Remove data ```semantifly delete data repo/README.md```

Add data source: ```semantifly add datasource $(gcloud config get-value project) --type gcp_project --label retail_prod --no_fetch```

Define new type: ```semantifly add type SuperCerealized supercerialized.proto```

Get data/metadata: ```semantifly describe data repo/README.md```

Type info: ```semantifly describe type architecture_diagram```

List data by type and data source: ```semantifly list data --type database --datasource aws_prod```

## Bulk Operations

Pull data from all configured data sources and individual records: ```semantifly pull --all --concurrency 20```

Pull data for specific data source: ```semantifly pull MyJiraInstance```

Generate embeddings: ```semantifly embed --all```

(Re)Generate embeddings for just one type (developer just added this type in a new data source, redefined/configure how a type should be embedded, etc.): ```semantifly embed type MyCustomType```

Export all gathered and indexed data to one file ```semantifly export myexport.semantifly --all```

Import data from semantifly export: ```semantifly import myexport.semantifly --all```

## RAG

CLI query: ```semantifly query "users database schema --max_records 10 --max_size 10KB --out queryresp.txt```

Start local query server: ```semantifly server start --port 8080```

Use with LLM: ```semantifly generate "sql to find first user to create record in users db" --max_records 2 --max_size 2KB --llm localhost:8080 --out oldest_user.sql```
