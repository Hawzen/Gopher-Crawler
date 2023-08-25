- Flask server
  - Using `facebook/bart-large-cnn` for summarization
  - Using TF-IDF for keyword extraction

The parallelization of the crawler is bottlenecked by the model inference. Splitting the two and running the model inference on a separate server would allow for more parallelization. I'm satisfied with the current performance, so I'm not going to do that.