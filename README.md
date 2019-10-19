# Golang URL Shortener
URL Shortener written in Go. Really fun project since I generally build APIs in Node.js and primarily use NoSQL databases. This task took me about ~5.5 hours.

## How-To
```shell
$ docker-compose up --build
```

Dependencies from the Docker image are MySQL and Golang.

## Tests
I didn't get a chance to write all the tests that I could as I wanted to stay as close to the 4 hour mark. I didn't get to approach it in a more idomatic way. When I started working on it, I realized I ended up setting it up as more of an end-to-end test instead of actual unit tests.

```shell
$ docker-compose -f docker-compose.test.yml up --build -d
$ go test -v
```

I also approached the tests in a Table Test way, so it would be more DRY.

## Endpoints 
There is a concept of the slug, where that's the randomized token that unlocks or redirects you to the original source.

**POST /api/shorten** Shorten's URLs
```shell
curl -X "POST" "http://localhost:3000/api/shorten" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "url": "http://localhost:3000/asdfdsf"
}'
```

**GET /api/stats?url=** Get Stats
```shell
curl "http://localhost:3000/api/stats?id={slug}"
```

**GET /{slug}** Redirect
```shell
curl "http://localhost:3000/{slug}"
```


## Assumptions
### General Application / Methodology
- I would generate hashes to act as the slug or transactional ID to append to the url (i.e. `http://localhost/${slug}`.)
- The slug would be randomly generated as a hash like component so that I can fulfill the following requirements: 
> 3. Is Unique; If a long url is added twice it should result in two different short urls.
> 4. Not easily discoverable; incrementing an already existing short url should have a low
probability of finding a working short url.
- Additionally, when a URL is requested to be shortened, I do not need to check if there is already a shortened version of it in the database because of the 3rd requirement.
- In addition this is _not_ production ready, and is definitely under the assumption that this is a POC/MVP.

### API
All the validation should occur at the handler level, that way the database doesn't get bogged down trying to figure out whether a URL is valid or not.

#### Shorten Endpoint
- If it's not a `POST` request respond with `405`
- If there's an error decoding the JSON respond with `500`
- If it's missing a `url` key and/or value is not a `url` respond with bad request
- Will not take a string longer than 128 characters

#### Stats Endpoint
- Checks by Query
- If value is not URL then let user know
- I am assuming that I do not have to provide these stats in real time, where if I had just requested for the stats and if additional increments/visits come in or if the 24 hour window has elasped, then it doesn't refresh the data for user.
- I'm assuming that I don't have to show historical output, and maybe just a number for the counts.
- If a user has a key that does not exist, still display 0 for returns
- Not assuming that I will always be expecting a valid URL, therefore case checking.

#### Redirect endpoint
- If a lookup is not found, then it would just default to a 404 status code. No increment or count will occur.

### Database
- To scale to a million requests - Ideally, I'd be using a cloud hosted solution (like AWS Aurora), that handles read replication for me.
- Since this application seems to be more read heavy, I'd ideally if I had more time... I would like to implement redis caching to ease the lookup for slug to destination.
- Would be nice to create a `trigger` to increment on read but I don't think that's possible.

## Thoughts
- It was a lot of fun to do. Initially, I had thought about doing this in Node.js and MongoDB... just because it was what I was comfortable with but I also wanted to challenge myself.
- I was also initially thinking about writing to a file on disk just because it would be easier to kick off and I wouldn't have to worry about any Docker aspects. Though, I hadn't really worked a lot with databases in Go.
- I also tried to avoid usage of too many third party packages and tried to focus only trying to use the standard library. However, there seemed to be some issues with validation of URLs via the recommended `url.ParseRequestURI`. ([Source](https://stackoverflow.com/questions/31480710/validate-url-with-standard-package-in-go))
- Need to work on better error handling, relied too much on JavaScript's `try/catch`.
