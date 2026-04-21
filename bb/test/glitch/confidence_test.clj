(ns glitch.confidence-test
  (:require [clojure.test :refer [deftest is testing]]
            [glitch.confidence :as conf]))

;; --- bayesian-combine ---

(deftest bayesian-combine-empty-test
  (testing "empty witnesses returns 0.0"
    (is (= 0.0 (conf/bayesian-combine [] 0.8)))))

(deftest bayesian-combine-single-test
  (testing "single witness: 1 - (1 - authority * base)"
    (let [result (conf/bayesian-combine [0.9] 0.8)]
      ;; 1 - (1 - 0.72) = 0.72
      (is (< (Math/abs (- result 0.72)) 0.001)))))

(deftest bayesian-combine-two-witnesses-test
  (testing "two witnesses combine higher than either alone"
    (let [result (conf/bayesian-combine [0.9 0.85] 0.8)]
      ;; 1 - (1-0.72)(1-0.68) = 1 - 0.28*0.32 = 1 - 0.0896 = 0.9104
      (is (> result 0.72))
      (is (< (Math/abs (- result 0.9104)) 0.001)))))

(deftest bayesian-combine-mixed-authorities-test
  (testing "mixed high/low authorities"
    (let [high (conf/bayesian-combine [0.9 0.9] 0.8)
          mixed (conf/bayesian-combine [0.9 0.3] 0.8)]
      ;; Higher authorities should yield higher combined confidence
      (is (> high mixed)))))

(deftest bayesian-combine-zero-base-test
  (testing "zero base confidence returns 0.0"
    (is (= 0.0 (conf/bayesian-combine [0.9 0.85] 0.0)))))

;; --- provider-authority ---

(deftest provider-authority-known-test
  (testing "known providers return their authority"
    (is (= 0.90 (conf/provider-authority "copilot")))
    (is (= 0.85 (conf/provider-authority "openrouter")))
    (is (= 0.60 (conf/provider-authority "lmstudio")))
    (is (= 0.55 (conf/provider-authority "ollama")))))

(deftest provider-authority-unknown-test
  (testing "unknown provider returns 0.50 fallback"
    (is (= 0.50 (conf/provider-authority "unknown-provider")))
    (is (= 0.50 (conf/provider-authority "my-custom-llm")))))

;; --- quality-score ---

(deftest quality-score-short-text-test
  (testing "short text yields lower score"
    (let [short-score (conf/quality-score 0.9 "ok")
          long-score  (conf/quality-score 0.9 (apply str (repeat 500 "x")))]
      (is (> long-score short-score)))))

(deftest quality-score-authority-matters-test
  (testing "higher authority yields higher score for same text"
    (let [high (conf/quality-score 0.9 "some response text")
          low  (conf/quality-score 0.3 "some response text")]
      (is (> high low)))))

(deftest quality-score-empty-text-test
  (testing "empty text gives log(1) * authority = 0"
    (is (= 0.0 (conf/quality-score 0.9 "")))))

;; --- harmonic-mean ---

(deftest harmonic-mean-all-ones-test
  (testing "all 1.0 values returns 1.0"
    (is (< (Math/abs (- 1.0 (conf/harmonic-mean [[1.0 1.0] [1.0 1.0] [1.0 1.0]]))) 0.001))))

(deftest harmonic-mean-near-zero-test
  (testing "one near-zero value pulls down the mean"
    (let [result (conf/harmonic-mean [[1.0 0.01] [1.0 1.0]])]
      (is (< result 0.1)))))

(deftest harmonic-mean-weighted-test
  (testing "weighted correctly — heavier weight on higher value biases up"
    (let [equal-weight   (conf/harmonic-mean [[1.0 0.5] [1.0 0.9]])
          heavy-on-high  (conf/harmonic-mean [[1.0 0.5] [3.0 0.9]])]
      (is (> heavy-on-high equal-weight)))))

(deftest harmonic-mean-empty-test
  (testing "empty scores returns 0.0"
    (is (= 0.0 (conf/harmonic-mean [])))))

(deftest harmonic-mean-all-zero-values-test
  (testing "all zero values returns 0.0 (filtered out)"
    (is (= 0.0 (conf/harmonic-mean [[1.0 0.0] [1.0 0.0]])))))

;; --- keyword-overlap ---

(deftest keyword-overlap-full-test
  (testing "full overlap returns 1.0"
    (is (= 1.0 (conf/keyword-overlap "deploy kubernetes cluster"
                                       "we will deploy the kubernetes cluster now")))))

(deftest keyword-overlap-none-test
  (testing "no overlap returns 0.0"
    (is (= 0.0 (conf/keyword-overlap "deploy kubernetes"
                                       "banana orange grape")))))

(deftest keyword-overlap-stopwords-test
  (testing "stopwords are excluded — only content words count"
    ;; "the" and "is" and "a" are stopwords
    (is (= 1.0 (conf/keyword-overlap "the big cat" "big cat sits")))))

(deftest keyword-overlap-empty-prompt-test
  (testing "empty prompt returns 1.0 (nothing to miss)"
    (is (= 1.0 (conf/keyword-overlap "" "any response")))))

(deftest keyword-overlap-partial-test
  (testing "partial overlap returns fraction"
    ;; prompt keywords: "deploy", "kubernetes", "cluster" (3 words)
    ;; response has "deploy" and "cluster" but not "kubernetes"
    (let [result (conf/keyword-overlap "deploy kubernetes cluster"
                                        "deploy the cluster")]
      (is (< (Math/abs (- result (/ 2.0 3.0))) 0.001)))))

;; --- domain-relevance ---

(deftest domain-relevance-on-topic-test
  (testing "on-topic preserves confidence (overlap*3 >= 1.0)"
    ;; Full overlap -> factor = min(1.0, 1.0*3) = 1.0
    (let [result (conf/domain-relevance 0.9 "deploy cluster" "deploy the cluster now")]
      (is (< (Math/abs (- result 0.9)) 0.001)))))

(deftest domain-relevance-off-topic-test
  (testing "off-topic penalizes confidence"
    (let [result (conf/domain-relevance 0.9 "deploy kubernetes cluster"
                                             "banana orange grape")]
      ;; overlap = 0.0, factor = 0.0
      (is (= 0.0 result)))))

(deftest domain-relevance-partial-test
  (testing "partial relevance scales confidence"
    ;; overlap ~0.33 -> factor = min(1.0, 0.33*3) ≈ 1.0
    ;; Actually overlap = 1/3, factor = min(1.0, 1.0) = 1.0
    ;; Need lower overlap for penalty to show
    (let [result (conf/domain-relevance 0.9 "alpha beta gamma delta epsilon"
                                             "alpha response")]
      ;; 1 out of 5 keywords -> overlap = 0.2, factor = 0.6
      (is (< result 0.9))
      (is (> result 0.0)))))

;; --- cosine-similarity ---

(deftest cosine-similarity-identical-test
  (testing "identical vectors return 1.0"
    (is (< (Math/abs (- 1.0 (conf/cosine-similarity [1 2 3] [1 2 3]))) 0.001))))

(deftest cosine-similarity-orthogonal-test
  (testing "orthogonal vectors return 0.0"
    (is (< (Math/abs (conf/cosine-similarity [1 0 0] [0 1 0])) 0.001))))

(deftest cosine-similarity-known-vectors-test
  (testing "known vector pair"
    (let [result (conf/cosine-similarity [1 0] [1 1])]
      ;; dot=1, mag-a=1, mag-b=sqrt(2), cos=1/sqrt(2)≈0.7071
      (is (< (Math/abs (- result (/ 1.0 (Math/sqrt 2.0)))) 0.001)))))

(deftest cosine-similarity-empty-test
  (testing "empty vectors return 0.0"
    (is (= 0.0 (conf/cosine-similarity [] [])))
    (is (= 0.0 (conf/cosine-similarity [1 2] [])))))

(deftest cosine-similarity-zero-magnitude-test
  (testing "zero vector returns 0.0"
    (is (= 0.0 (conf/cosine-similarity [0 0 0] [1 2 3])))))
