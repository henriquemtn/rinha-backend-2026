Coisas que aprendi a fazer com o projeto:

- Como construir e publicar imagem no Docker Hub (build, tag, push).
- Diferenca entre build local e usar imagem pronta no docker-compose.
- Como o docker-compose usa limites de CPU/memoria e o que isso muda no runtime.
- Como usar nginx como load balancer com duas instancias da API.
- Como usar socket Unix entre nginx e API para reduzir overhead.
- Como entender erros comuns: 502/503, OOM (exit 137), conflicts de containers.
- Como rodar benchmark local e interpretar p50/p95/p99 e timeouts.
- Como configurar variaveis de ambiente no container para ajustar o modelo.

Coisas que preciso estudar mais a fundo:

- Go: organizacao de projeto, pacotes e layout recomendado.
- Go: performance (allocs, parsing JSON sem refletir, profiling, pprof).
- Go: concorrencia e limites de throughput (fila, backpressure, timeouts).
- Algoritmos de busca vetorial (IVF, escolha de k, n_probe, tradeoffs).
- Como calibrar latencia vs acuracia sem quebrar o contrato do desafio.
- Como montar uma rotina de testes locais parecida com o teste da rinha.


14/05/2026

Ontem quando submeti o segundo teste acabei piorando o resultado, então decidi olhar outros projetos, ver o que a galera tava comentando sobre no discord/linkedin e percebi que a minha Api ainda estava muito Crua, decidi entender como as outras pessoas estavam resolvendo o problema e re-estruturando a minha api.. documentei tudo no REFACTOR.md