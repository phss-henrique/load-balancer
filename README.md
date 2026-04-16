# Go Load Balancer & OS Stress Test Analysis

Este projeto é uma implementação de um **Load Balancer HTTP (Proxy Reverso)** escrito em Go (Golang), utilizando o algoritmo de roteamento **Round Robin**. 

Além do desenvolvimento da arquitetura de roteamento distribuído, este repositório documenta um experimento prático de testes de carga extremos utilizando a ferramenta **Vegeta**, analisando como o limite de conexões TCP do Windows (Ephemeral Ports) se comporta sob estresse severo em hardware de alta performance.

## Arquitetura do Projeto

A infraestrutura local consiste em aplicações completamente independentes comunicando-se via rede:

* **Load Balancer (`:8080`):** Intercepta o tráfego HTTP de entrada, aplica a lógica matemática do Round Robin com concorrência segura (`sync/atomic`) e realiza o proxy reverso.
* **Back-ends (`:8081`, `:8082`, `:8083`):** Três instâncias de servidores de destino rodando simultaneamente através de *Goroutines*, simulando microsserviços.

## Benchmarks e Testes de Carga (Stress Testing)

Os testes foram desenhados para encontrar o gargalo da aplicação. 

**Hardware utilizado no teste:**
* **CPU:** Intel Xeon E5-2680 v4 (14 Núcleos Físicos / 28 Threads)
* **OS:** Windows 
* **Ferramenta:** Vegeta (HTTP load testing tool)

### 🟢 Teste 1: Tráfego Sustentável (Baseline)
O primeiro teste validou o funcionamento da arquitetura sob uma carga considerável, mas realista.

* **Carga:** 1.000 requisições por segundo durante 5 segundos.
* **Resultado:** **100% de Sucesso** (5.000 requisições processadas).
* **Latência (p50):** ~0.53ms
* **Latência (p99):** ~7.47ms

**Análise:** O gerenciamento de *Goroutines* do Go distribuiu a carga perfeitamente pelos 28 núcleos lógicos do Xeon. A aplicação roteou o tráfego em pura memória com latências sub-milissegundo, sem gerar estresse na CPU.

### 🔴 Teste 2: O Colapso do Sistema Operacional (Port Exhaustion)
O segundo teste buscou o limite absoluto do ambiente local, tentando exaurir os recursos da máquina.

* **Carga:** 15.000 requisições por segundo durante 10 segundos.
* **Resultado do Vegeta:**
  * **Success Ratio:** `12.89%`
  * **Status Codes:** `0:118815`, `200:17576`
  * **Error:** `dial tcp 0.0.0.0:0->127.0.0.1:8080: connectex: No connection could be made because the target machine actively refused it.`

**Análise do Gargalo (Post-mortem):**
O gargalo **não** foi o processador Xeon, e sim o limite arquitetural do Sistema Operacional. O erro ocorreu devido à **Exaustão de Portas Efêmeras (Ephemeral Port Exhaustion)**. 

Ao tentar abrir 15.000 conexões TCP por segundo localmente, o limite nativo de alocação de portas dinâmicas do Windows foi atingido rapidamente. Sem portas disponíveis para abrir novos *sockets*, o Windows passou a recusar ativamente as novas conexões (gerando os códigos de status `0` no Vegeta), antes mesmo que o tráfego chegasse à camada da aplicação em Go. 

Isso demonstra na prática um cenário de DDoS (Negação de Serviço) focado na exaustão de recursos de rede, evidenciando por que ambientes de produção (geralmente Linux em Data Centers) exigem ajustes finos no Kernel (`sysctl`, tuning de TCP/IP) para sustentar volumes massivos de conexões simultâneas.

## Como rodar o projeto

1. Clone o repositório.
2. Abra dois terminais na raiz do projeto.
3. No Terminal 1, inicie as APIs de destino:
 
   ```bash
   go run cmd/api/backends.go
   ```
4. No Terminal 2, inicie o Load Balancer:
  
   ```bash
   go run cmd/proxy/main.go
   ```
5. Acesse http://localhost:8080 no navegador ou via curl para observar o balanceamento de carga em ação.
