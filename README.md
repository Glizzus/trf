# Ministry of Truth

The ministry is the service responsible for spoofing articles.

It is named for the Ministry of Truth in George Orwell's novel 1984.

## Environment Variables

| Name | Description | Values | Default | Required | Depends On |
|------|-------------|--------|---------|----------|------------|
| `MINISTRY_SPOOFER_TYPE` | The type of spoofer to use | `OPENAI`, `MOCK` | `MOCK` | Yes | |
| `MINISTRY_SPOOFER_OPENAI_API_KEY` | The OpenAI API key | | | Conditional | `MINISTRY_SPOOFER_TYPE=OPENAI` |
| `MINISTRY_PROMPT_TYPE` | The type of prompt to use. | `FILE`, `STATIC` | `STATIC` | Yes | |
| `MINISTRY_PROMPT_FILE` | The file containing the prompt. | | | No | |

## Prompt Templates

When creating custom prompts, there are two required templates: `system.tmpl` and `user.tmpl`.

### `system.tmpl`

This template is used to generate the system's directive to the AI. Currently, no variables are passed to this template.

Example:

```
You are the Ministry of Truth. You will be given an article, as well as a rating. You must write a parody of the article.
```

### `user.tmpl`

This template is used to template the article to feed to the AI. There are two variables; the first is the content, and the second is the rating.

Example:

```
The following article is rated {{.Rating}}:

{{.Content}}

Please write a parody of the article.
```


## Endpoints

### `GET /health`

Returns a 200 status code if the service is healthy.

### `POST /spoof`

Spoofs an article.

#### Request

```json
{
  "content": "The quick brown fox jumps over the lazy dog.",
  "rating": "True"
}
```

#### Response

```json
{
  "content": "The quick brown fox never jumped over the lazy dog."
}
```