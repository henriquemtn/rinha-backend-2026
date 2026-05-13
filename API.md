# API - Descricao funcional

Este documento descreve o comportamento da API e sua logica de pontuacao.

## Inicializacao

- Define `GOMAXPROCS=1` para respeitar o limite de CPU por container.
- Carrega recursos em memoria:
  - `normalization.json` (limites e fatores de normalizacao).
  - `mcc_risk.json` (risco por MCC).
  - `references.json.gz` (vetores e labels de referencia).
- Exibe estatisticas de carga (tempo, quantidade de referencias e memoria).
- Sobe servidor HTTP em `LISTEN_ADDR` (padrao `:9999`).
- Se `LISTEN_ADDR` comeca com `/`, usa socket Unix e ajusta permissao para leitura/escrita.

## Endpoints

### GET /ready

Healthcheck simples. Retorna `200` sem corpo.

### POST /fraud-score

Recebe um payload JSON e retorna uma decisao de aprovacao e a pontuacao de fraude.

Exemplo de resposta:

```json
{
  "approved": true,
  "fraud_score": 0.2
}
```

Erros:

- `405` se metodo nao for POST.
- `400` para JSON invalido ou payload invalido.
- `500` se os recursos nao foram carregados.

## Estrutura do payload

- `id`: string
- `transaction`:
  - `amount`: float
  - `installments`: int
  - `requested_at`: RFC3339
- `customer`:
  - `avg_amount`: float
  - `tx_count_24h`: int
  - `known_merchants`: string[]
- `merchant`:
  - `id`: string
  - `mcc`: string
  - `avg_amount`: float
- `terminal`:
  - `is_online`: bool
  - `card_present`: bool
  - `km_from_home`: float
- `last_transaction` (opcional):
  - `timestamp`: RFC3339
  - `km_from_current`: float

## Fluxo de pontuacao (KNN)

- Gera um vetor numerico de 14 dimensoes a partir do payload.
- Calcula a distancia quadratica para todas as referencias carregadas.
- Mantem os `k=5` vizinhos mais proximos.
- Calcula `fraud_score = fraud_count / k`.
- Aprova se `fraud_score < 0.6`.

## Vetorizacao (14 dimensoes)

Dimensoes e regras:

1. `amount` normalizado por `max_amount`.
2. `installments` normalizado por `max_installments`.
3. Razao `amount / avg_amount` normalizada por `amount_vs_avg_ratio`.
4. Hora do dia (`0-23` / 23).
5. Dia da semana (segunda=0, domingo=6 / 6).
6. Minutos desde a ultima transacao (ou `-1` se ausente).
7. Km da ultima transacao ate a atual (ou `-1` se ausente).
8. Km do terminal ate a casa (normalizado).
9. Transacoes nas ultimas 24h (normalizado).
10. Terminal online (`1` ou `0`).
11. Cartao presente (`1` ou `0`).
12. Merchant desconhecido (`1` se nao estiver em `known_merchants`).
13. Risco por MCC (default 0.5 se ausente).
14. Media de valor do merchant (normalizada).

## Normalizacao e riscos

- `normalization.json` define limites maximos para cada feature numerica.
- `mcc_risk.json` define o risco por MCC; quando ausente, usa 0.5.

## Referencias e quantizacao

- Cada referencia possui `vector[14]` e `label` (fraud/legit).
- Vetores de referencia sao quantizados para `uint16` no carregamento.
- Na comparacao, os valores sao desquantizados para `float64`.

## Observacoes de tempo

- Datas sao parseadas com `RFC3339`.
- `last_transaction` opcional altera duas features (tempo e distancia).
