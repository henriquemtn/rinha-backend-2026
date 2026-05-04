# 🐔 Rinha de Backend 2026 — Fraud Detection (Go)

Implementação da Rinha de Backend 2026 utilizando Golang, com foco em alta performance, eficiência de memória e detecção de fraude via busca vetorial.

Este projeto tem como principal objetivo aprendizado e evolução em backend, explorando na prática conceitos como processamento de alto volume, otimização de performance e arquitetura sob restrições.

A Rinha está sendo usada aqui mais como um ambiente de experimentação e aprofundamento técnico, especialmente no meu primeiro contato mais sério com Go, do que como uma busca puramente competitiva por ranking.

## 🚀 Stacks

[![My Skills](https://skillicons.dev/icons?i=go,docker,nginx)](https://skillicons.dev)

* **Go (net/http)** — servidor leve e performático
* **Docker** — conteinerização
* **Nginx** — load balancer (round-robin)
* **KNN (brute force otimizado)** — detecção de fraude

## 📡 Endpoints

### `GET /ready`

Verifica se a aplicação está pronta.

```bash
curl http://localhost:9999/ready
```

---

### `POST /fraud-score`

Recebe uma transação e retorna o score de fraude.

```json
{
  "approved": false,
  "fraud_score": 0.8
}
```

## 📁 Estrutura

```bash
.
├── api/
│   ├── main.go
│   ├── handlers.go
│   ├── model.go
│   ├── service.go
│   └── vector.go
├── docker-compose.yml
└── README.md
```

## 🏁 Status

🚧 Em desenvolvimento e aprendendo com o processo!

---