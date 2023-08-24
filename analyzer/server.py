from flask import Flask, request
from bs4 import BeautifulSoup
app = Flask(__name__)


# Summarization
@app.route('/summarize', methods=['POST'])
def summarize():
    html = request.data.decode('utf-8')
    if html == '':
        return 'No html provided'
    text = get_text_from_html(html)
    return summarize_text(text)

def summarize_text(text: str) -> str:
    # stub
    return text


# Keywords
@app.route('/keywords', methods=['POST'])
def keywords():
    """stub"""
    html = request.data.decode('utf-8')
    text = get_text_from_html(html)
    return get_keywords(text)

def get_keywords(text: str) -> list:
    """stub"""
    return ["apple", "banana", "orange"]


# Utils
def get_text_from_html(html: str):
    soup = BeautifulSoup(html, 'html.parser')
    return soup.get_text()


if __name__ == '__main__':
    app.run(port=6969)