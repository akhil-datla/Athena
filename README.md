# Athena
Athena is the Q&A search engine built for your own FAQs and answers. This means that Athena uses your existing question and answer database and enables fuzzy search queries on the questions. Athena ranks queries against the questions using Levenshtein distance. When running Athena, you will need to pass an argument for your CSV file that contains the questions and answers. The CSV file should be formatted as `question,answer`: 

```
Who made Athena?,Akhil Datla
What language is Athena made in?,Go
```
Screenshot of Athena's Search API:
<img width="859" alt="Screen Shot 2022-05-14 at 4 21 12 PM" src="https://user-images.githubusercontent.com/66145155/168451072-9d38c4dc-9e83-4153-aa02-82f81a3aa30d.png">
