## Requisitos

- Go 1.25 ou superior
- Os dois jogadores precisam estar em rede acessível um ao outro (mesma LAN ou loopback) e conhecer o IP/porta um do outro
- Cada peer escuta numa porta UDP local; portas devem estar liberadas no firewall

## Compilação

```sh
go build -o naval ./cmd
```

Para rodar direto sem gerar binário:

```sh
go run ./cmd --listen :9000
```

## Uso

Em cada máquina (ou em duas portas diferentes da mesma máquina), execute:

```sh
./naval --listen :9000
```

A flag `--listen` define a porta UDP local. O programa pergunta:

```
(C)onvidar ou (E)sperar?
```

- **C (Convidar):** digite `IP:porta` do oponente; o convite é reenviado a cada 1s até ser aceito. Quem convida atira primeiro.
- **E (Esperar):** aguarda receber um convite. Ao chegar, pergunta `aceitar? (s/n)`.

Durante o jogo, os tabuleiros aparecem lado a lado:

```
   A B C D E F G H I J       A B C D E F G H I J
 1 . . . . . . . . . .     1 . . . . . . . . . .
 2 . . S . . . . . . .     2 . . . . . . X . . .
...
   tabuleiro: voce            tabuleiro: oponente
```

Símbolos:

- `.` água / desconhecido
- `S` seu navio
- `X` acerto (no seu tabuleiro: o oponente acertou; no do oponente: você acertou)
- `O` o oponente errou no seu tabuleiro
- `~` você errou no tabuleiro do oponente

Para atirar, digite a coordenada estilo batalha naval (coluna `A`–`J` + linha `1`–`10`):

```
B7
```

Quem acerta atira de novo; quem erra passa a vez. O jogo termina quando todos os navios de um dos lados são afundados.

## Teste rápido em loopback

Em dois terminais:

```sh
# terminal 1
go run ./cmd --listen :9000
# escolher: E

# terminal 2
go run ./cmd --listen :9001
# escolher: C
# IP:porta do oponente: 127.0.0.1:9000
```
