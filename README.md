# System Monitorowania i Analizy Wydatków na Subskrypcje

Projekt zaliczeniowy z przedmiotu Inżynieria Oprogramowania (Politechnika Śląska, AEI, Semestr 5, 2026/2027).

Aplikacja webowa do monitorowania i analizy subskrypcji (np. Netflix, Spotify, chmura): pokazuje ile nas to kosztuje miesięcznie i rocznie, przypomnienia przed płatnością, historia płatności i podsumowania w jednej walucie.

## Zespół

- Piotr Krzywoń - Lider zespołu; frontend: lista i formularz subskrypcji, historia płatności; backend: historia płatności i eksport CSV
- Adrian Hevko - frontend: podstawa aplikacji, logowanie, podsumowanie kosztów; backend: waluty i obliczanie kosztów; tablica projektu
- Kamil Hibszer - backend: serwer, model danych, repozytoria, logowanie, API subskrypcji, Worker

## Technologie (plan)

- backend: Go (biblioteka standardowa)
- frontend: React + TypeScript (Vite)

## Organizacja

Zadania i postęp prac: zakładka **Projects** tego repozytorium.
Dokumentacja etapów oddawana jest na PZE.

## Uruchomienie

### Backend (Go 1.27.3)

```bash
cd backend
go run ./cmd/server
```

### Frontend (Node.js)

```bash
cd frontend
npm install
npm run dev
```

## Diagramy

```bash
cd diagrams
python generate.py
```

> [!WARNING]
> Należy pobrać **plantuml.jar** ze strony https://plantuml.com/download oraz umieścić go w katalogu `diagrams/`

- [Diagram architektury systemu](diagrams/out/01_architektura.png) · [Źródło Graphviz](diagrams/01_architektura.dot)
- [Diagram klas domeny](diagrams/out/02_model_domenowy.png) · [Źródło PlantUML](diagrams/02_model_domenowy.puml)
- [Diagram warstwy logiki biznesowej](diagrams/out/03_warstwa_logiki.png) · [Źródło PlantUML](diagrams/03_warstwa_logiki.puml)
- [Diagram warstwy danych](diagrams/out/04_warstwa_danych.png) · [Źródło PlantUML](diagrams/04_warstwa_danych.puml)