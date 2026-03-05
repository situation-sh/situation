---
title: Overview
summary: How to consume collected data
order: 1
---

Situation is a bit special since it exposes an **SQL API**. 

This is different from the common REST API we consume across the web. 
Just grab the SQL schema [like you will do with an OpenAPI spec] and you are done!

As we support both SQLite and PostgreSQL, we list the corresponding SQL tables in the next pages (only the types may differ).

## SDK

Can we automatically generate clients? Yes of course. Currently, we generate:

- +logos:python+ **python** package relying on [SQLModel](https://sqlmodel.tiangolo.com/) 
- +logos:typescript-icon+ **typescript** package using [drizzle](https://orm.drizzle.team/).

Currently, these SDK are not distributed through package managers. You should grab them manually from the [dedicated Github action]({{ latest_sdk_workflow_url }}).