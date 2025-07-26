# Riot Team Service (League of Legends API Helper)

Este projeto é um backend em Go que permite criar times com jogadores do LoL e armazená-los em um banco PostgreSQL. No futuro, ele será expandido para calcular estatísticas como win rate dos times.

---

## 🚀 Tecnologias utilizadas

- Go (Golang)
- Gin (framework HTTP)
- SQLX (acesso ao banco)
- PostgreSQL
- Docker + Docker Compose

---

## 📁 Estrutura do Projeto

- `cmd/` – entrada da aplicação (`main.go`)
- `internal/adapter/` – handlers HTTP e repositórios de banco
- `internal/usecase/` – regras de negócio
- `internal/domain/` – entidades e interfaces
- `internal/util/` – utilitários para o código
- `configs/` – configuração de banco
- `migrations/` – arquivos SQL para criação das tabelas

---

## 🔴️ Necessário criar o .env.dev!!

Por segurança não foi colocado o arquivo .env.dev no projeto, ele deve conter a variável
responsável por armazenar a chave de acesso para o portal de desenvolvimento da API da Riot.

Adicione esse arquivo na pasta raiz, e digite o seguinte: `RIOT_API_KEY=SENHA`

Substitua **SENHA** pela chave do portal da API, você consegue ela no [Riot Developer Portal](https://developer.riotgames.com/)

---

## 🐘 Subindo o banco com Docker

Suba o banco PostgreSQL com o seguinte comando:

```bash
docker-compose up -d
```

Isso iniciará um contêiner PostgreSQL com as seguintes configurações:

- **Banco**: `riot_db`
- **Usuário**: `user`
- **Senha**: `password`
- **Porta**: `5432`

---

## 🛠 Rodando as migrations

Para aplicar os scripts SQL do projeto, você precisa da CLI do [`golang-migrate`](https://github.com/golang-migrate/migrate).

Após instalar, rode:

```bash
migrate -path ./migrations -database "postgres://user:password@localhost:5432/riot_db?sslmode=disable" up
```

> Isso criará as tabelas `team` e `player` no banco.

---

## 🧲 Testando o endpoint

### Criar um time com jogadores

- **Método**: `POST`
- **URL**: `http://localhost:8080/teams`
- **Body (JSON)**:

```json
{
  "name": "Gaplandia",
  "players": [
    { "gamer_name": "phini", "tag_line": "br1" },
    { "gamer_name": "hampstead", "tag_line": "pixys" },
    { "gamer_name": "thumpy", "tag_line": "pixys" },
    { "gamer_name": "debocholandia", "tag_line": "br1" },
    { "gamer_name": "bebe reborn", "tag_line": "unhe" }
  ]
}
```

- **Resposta esperada (HTTP 201)**:

```json
{
  "message": "Team created successfully"
}
```

---

## ✅🚀 Próximos passos

- Integração com API da Riot Games ✔️
- Cálculo de win rate dos jogadores ️️ ✔️
- Implementar buscas concorrentes ️️ ✔️
- Ranqueamento de times
- Autenticação e controle de acesso

