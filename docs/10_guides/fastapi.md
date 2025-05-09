---
title: FastAPI
summary: Example with http backend
order: 30
---

In this guide, we build a backend able to receive the payload from Situation. This backend is a simple FastAPI application that will print the payload to stdout.

## Requirements

Initialize a project with [uv](https://docs.astral.sh/uv/) for instance.

```bash
uv init situation-fastapi
cd situation-fastapi
```

Add some dependencies:

```bash
uv add 'fastapi[standard]' 
```

## First steps

Create a file `main.py` with the following content:

```python
# main.py
from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/", response_class=HTMLResponse)
async def root():
    return """
    <html>
        <body>
            <div>Everything works fine</div>
        </body>
    </html>"""


@app.post("/")
async def situation(payload: dict):
    print(payload)
    return
```

You can check that everything works fine by running the server:

```bash
uv run fastapi dev 
```

Then, open your browser and go to `http://127.0.0.1:8000/`.

Now, let us try to send data to this backend. On a new terminal, get the latest binary:

```bash
curl -sLo ./situation https://github.com/situation-sh/situation/releases/download/{{ latest_tag }}/situation-{{ latest_version }}-amd64-linux
chmod +x ./situation
```

and run it with the following command:

```bash
./situation --http --http-url=http://127.0.0.1:8000/
```

On the server, you should see the payload printed in the terminal.


## Security 

You can add some security to the endpoint by authorizing only our agent. By default the agent sends its id in the `Authorization` header. 

To ensure you have a unique id, you can use the `refresh-id` command to regenerate it.

```bash
./situation refresh-id # regenerate an id
./situation id > .authorized_id # store the id to a local file
```

Now restrict the access to the endpoint:
```python
from fastapi import Depends, FastAPI
from fastapi.security import APIKeyHeader

app = FastAPI()

authorized_id = open(".authorized_id").read().strip(" \n")
auth = APIKeyHeader(name="authorization")

# ...

@app.post("/")
async def situation(payload: dict, agent_id: str = Depends(auth)):
    if agent_id != authorized_id:
        raise HTTPException(status_code=403, detail="Not authenticated")
    print(payload)
    print(agent_id)
    return
```

You can check that the endpoint is now protected.

```bash
curl -X POST \
    -H 'Authorization: 00000000-0000-0000-0000-000000000000' \
    -H 'Content-Type: application/json' \
    --data '{}' \
    'http://127.0.0.1:8000/'
```


## Payload integrity

Currently, the endpoint can accept any payload. Fortunately, the agent comes with a json-schema that describes what it sends. We will use it to validate what we receive.


The idea is to create a `pydantic` model that will validate the payload. To turn the json-schema into a pydantic model, we can use [`datamodel-code-generator`](https://github.com/koxudaxi/datamodel-code-generator).

```bash
uv add --group dev datamodel-code-generator
```

Then, we can generate the model (`models.py`) with the following command:

```bash
uv run datamodel-codegen \
    --url 'https://github.com/situation-sh/situation/releases/download/{{ latest_tag }}/schema.json' \
    --input-file-type jsonschema \
    --output=models.py \
    --output-model-type pydantic_v2.BaseModel \
    --use-annotated \
    --use-union-operator \
    --use-field-description \
    --reuse-model \
    --collapse-root-models
```

Instead of using `dict` we now have fully parsed model.

```python
from models import Payload
# ...
@app.post("/")
async def situation(payload: Payload, agent_id: str = Depends(auth)):
    if agent_id != authorized_id:
        raise HTTPException(status_code=403, detail="Not authenticated")
    print(payload)
    print(agent_id)
    return
```

You can re-run the agent and also check that you can't send anything to the endpoint. The following request returns a 422 error.

```bash
curl -X POST \
    -H "Authorization: $(cat .authorized_id)" \
    -H 'Content-Type: application/json' \
    --data '{"foo": "bar"}' \
    'http://127.0.0.1:8000/'
```

## Going further 

Finally, you are free to use the data as you want. In this last example, we will list all the received records in a table.

We start by creating a template for the HTML page. We will use [Jinja2](https://jinja.palletsprojects.com/) to render the page.

```bash
mkdir -p templates 
touch templates/index.html
```

The content of the template is:

```html
<!DOCTYPE html>
<html lang="en">

<head>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link
        href="https://fonts.googleapis.com/css2?family=Geist+Mono:wght@100..900&family=Geist:wght@100..900&display=swap"
        rel="stylesheet">
    <style>
        body {
            width: 100vw;
            height: 100vh;
            overflow: hidden;
            font-family: "Geist", sans-serif;
        }

        article {
            max-width: 1200px;
            margin: 0 auto;
            width: 100%;
        }

        h1 {
            font-size: 2.5rem;
            font-weight: 800;
        }

        .container {
            width: 100%;
            height: 100%;
            display: flex;
            flex-direction: column;
        }

        .table-wrapper {
            width: 100%;
            height: 100%;
            max-height: 70vh;
            overflow: auto;
            font-family: "Geist Mono", monospace;
            font-size: small;
        }

        table {
            border-collapse: collapse;
        }

        thead th {
            height: 1.5rem;
            border-bottom: 1px solid lightgray;
            min-width: 75px;
            text-align: left;
        }

        td {
            vertical-align: middle;
            padding: 0 2px;
        }
    </style>
</head>

<body>
    <article>
        <div class="container">
            <h1>Situation x FastAPI</h1>
            <div class="table-wrapper">
                <table style="width: 100%">
                    <thead>
                        <tr>
                            <th>Timestamp</th>
                            <th>Agent</th>
                            <th>Machines</th>
                            <th>Duration</th>
                            <th>Errors</th>
                        </tr>
                    </thead>
                    <tbody>
                        {% raw %}{% for record in records %}
                        <tr>
                            <td>{{ record.extra.timestamp }}</td>
                            <td>{{ record.extra.agent }}</td>
                            <td>{{ record.machines|length }}</td>
                            <td>{{ "%.3f"|format(record.extra.duration / 1e9) }}s</td>
                            <td>
                                <ul>
                                    {% for error in record.extra.errors %}
                                    <li><b>{{ error.module }}</b>: {{error.message}}</li>
                                    {% endfor %}
                                </ul>
                            </td>
                        </tr>
                        {% endfor %}{% endraw %}
                    </tbody>
                </table>
            </div>
        </div>
    </article>
</body>

</html>
```

Here is the update of `main.py`:

```python
from typing import List

from fastapi import Depends, FastAPI, HTTPException, Request
from fastapi.responses import HTMLResponse
from fastapi.security import APIKeyHeader
from fastapi.templating import Jinja2Templates

from models import Payload

app = FastAPI()

templates = Jinja2Templates(directory="templates")

authorized_id = open(".authorized_id").read().strip(" \n")
auth = APIKeyHeader(name="Authorization")

# store all the payloads in memory
records: List[Payload] = []


@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    return templates.TemplateResponse(
        request=request,
        name="index.html",
        context={"records": records},
    )


@app.post("/")
async def situation(payload: Payload, agent_id: str = Depends(auth)):
    if agent_id != authorized_id:
        raise HTTPException(status_code=403, detail="Not authenticated")
    records.append(payload)
    return
```