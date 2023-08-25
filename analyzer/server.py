from bs4 import BeautifulSoup
from flask import Flask, request
from sklearn.feature_extraction.text import TfidfVectorizer
from transformers import pipeline

app = Flask(__name__)

summarizer = pipeline('summarization', model='facebook/bart-large-cnn')

# Summarization
@app.route('/summarize', methods=['POST'])
def summarize():
    html = request.data.decode('utf-8')
    if html == '':
        return 'No html provided'
    text = get_text_from_html(html)

    return summarize_text(text)


def summarize_text(text: str) -> str:
    text = text.replace('\n', ' ')
    # Limit to 1000 characters. This is a limitation of the summarization model. Use a different model if you need more
    text = text[:1000]
    return summarizer(text, max_length=500, min_length=30, do_sample=True)[0]['summary_text']


# Keywords
@app.route('/keywords', methods=['POST'])
def keywords():
    html = request.data.decode('utf-8')
    text = get_text_from_html(html)
    return get_keywords(text)


def get_keywords(text: str, num_keywords: int = 10) -> list:
    vectorizer = TfidfVectorizer(max_features=num_keywords)
    try:
        vectorizer.fit_transform([text])
    except ValueError:
        return []
    return vectorizer.get_feature_names_out().tolist()


# Utils
def get_text_from_html(html: str):
    soup = BeautifulSoup(html, 'html.parser')
    return soup.get_text()

if __name__ == '__main__':
    app.run(port=9898)
