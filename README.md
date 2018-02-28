# Oppgave fagintervju

Filen data.csv inneholder loggdata for innloggede brukere på et nettsted som inneholder 3 sider for henholdsvis biler, båter og sykler. Hver rad inneholder informasjon om brukeren, hvilken side som ble besøkt og hvor lenge brukeren ble på siden målt i millisekunder. Alle data er generert og inneholder tilfeldig informasjon.

## Komme i gang

Fork dette repository og sjekk det ut på din maskin.

Docker må installeres og ligge tilgjengelig på kommandolinje. Docker kan lastes ned til ulike operativsystem fra: [https://www.docker.com/community-edition](https://www.docker.com/community-edition "Docker community edition"), versjon bør være minimum 17.05. Dersom en annen utgave benyttes, må den konfigureres enten slik at nettverket til containere i docker er tilgjengelig på localhost eller endre ip adresser for publiserte tjenester og tilkoblinger mot dem. Følgende kommando kan kjøres for å verifisere at docker kjører korrekt: `docker run hello-world`

Start infrastruktur ved å kjøre: `docker-compose up -d`

Stopp infrastruktur ved å kjøre: `docker-compose down`. Ved stopping av infrastruktur, slettes samtidig alle data. Data kan gjøres persistente ved å legge et docker volum til kafka og elasticsearch containerne og konfigurere data fra dem til å lagres på volumene.

Se hvilke containere som kjører: `docker ps`, her skal det listes 3 tjenester: kafka, elasticsearch og kibana. Dersom en av tjenestene ikke kjører må man feilsøke årsaken til at den ikke starter, kommandoer for å sjekke logger:

```bash
docker logs --tail 2000 -f kafka
docker logs --tail 2000 -f elasticsearch
docker logs --tail 2000 -f kibana
```

## Oppgaver

Oppgavene innebærer arbeid på selvstendige mikrotjenester. De tre mikrotjenesten har hver sin mappe: parse-file, index-documents og rest-api. Implementasjonsspråk er valgfritt, det anbefales å benytte et språk som har klientbiblioteker til Kafka (versjon 0.10.2) [https://cwiki.apache.org/confluence/display/KAFKA/Clients](https://cwiki.apache.org/confluence/display/KAFKA/Clients "Kafka client libraries") og Elasticsearch (versjon 6.x) [https://www.elastic.co/guide/en/elasticsearch/client/index.html](https://www.elastic.co/guide/en/elasticsearch/client/index.html "Elasticsearch official clients"), [https://www.elastic.co/guide/en/elasticsearch/client/community/current/index.html](https://www.elastic.co/guide/en/elasticsearch/client/community/current/index.html "Elasticsearch community clients").

Implementasjoner skal som minimum inneholde dokumentasjon i README.md fil (Eksisterer allerede i mappene) for hvordan man henter avhengigheter, bygger og kjører tjenesten. Tjenestene skal kjøre som docker containere, og det skal fortsettes på Dockerfile som er lagt inn. Velg et baseimage som er riktig for implementasjonsspråk, noen forslag er allerede lagt inn i Dockerfile, men man står helt fritt til å velge ønsket baseimage.

1. Filen data.csv skal parses, og hver linje skal leveres som et json dokument på en kafka topic ved navn "pageviews". Kolonnenavn angis i første rad av filen og dette benyttes også til navn på properties i json dokumenter. Applikasjonen skal leveres som en mikrotjeneste i valgfritt språk, implementasjon legges i mappen parse-file.

2. JSON dokumenter på kafka topic "pageviews" skal hentes ut fra Kafka og indekseres i elasticsearch i en indeks kalt "views". Opprett et indekspattern i kibana for "views", det er ikke noe dato/tidspunkt felt i data, indeks pattern må dermed opprettes uten en tidspunkt kolonne. Applikasjonen skal leveres som en mikrotjeneste i valgfritt språk, implementasjon legges i mappen index-documents.

3. Det skal tilbys et REST api for å hente ut aggregert informasjon om sidevisninger. APIet skal kun tilby spørringer over HTTP med GET, og skal levere json dokumenter som resultat. Endepunkter som skal implementeres ligger under. Applikasjonen skal eksponere apiet over port 8904, eventuelt endres dokumentasjonen dersom man eksponerer over en annen port. Applikasjonen skal leveres som en mikrotjeneste i valgfritt språk, implementasjon legges i mappen rest-api.

GET http://localhost:8904/api/views/total/ -> Skal returnere antall visninger per side, sortert på antall visninger, eksempel:

```json
[
    {
        "page": "http://testsite.zz/98/bikes",
        "views": 73
    },
    {
        "page": "http://testsite.zz/43/cars",
        "views": 39
    }
]
```

GET http://localhost:8904/api/views/total-time-viewed/ -> Skal returnere total visningstid per side i hele sekunder, sortert på total visningstid, eksempel:

```json
[
    {
        "page": "http://testsite.zz/57/boats",
        "totalTimeViewed": 65718
    },
    {
        "page": "http://testsite.zz/43/cars",
        "totalTimeViewed": 51487
    }
]
```

GET http://localhost:8904/api/customers/ -> Skal returnere de 5 brukerne som har tilbragt mest tid på nettstedet med totalt antall visninger og total visningstid i hele sekunder, sortert på total visningstid, eksempel:

```json
[
    {
        "customerId": "440336c5-293a-488a-a829-79a1f3ea00ef",
        "views": 27,
        "totalTimeViewed": 17543
    },
    {
        "customerId": "933af785-dfc4-404a-8ca9-be73e78c5db0",
        "views": 31,
        "totalTimeViewed": 13526
    },
    {
        "customerId": "30d452b9-a703-467c-a067-dae10372ebc0",
        "views": 15,
        "totalTimeViewed": 9746
    },
    {
        "customerId": "254b5f75-be85-49b3-9bee-89d31912d6a0",
        "views": 12,
        "totalTimeViewed": 5961
    },
    {
        "customerId": "79535db0-83b6-4fc5-9aee-aac517230877",
        "views": 16,
        "totalTimeViewed": 5952
    }
]
```

4. Import av filen til elasticsearch er ikke idempotent, det vil si, kjører man filen på nytt vil man få duplikater, mens man ved idempotent prosessering aldri ville sett en endring uansett hvor mange ganger man reimporterer data. Gjør dataimport idempotent.
