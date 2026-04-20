(ns glitch.provider-test
  (:require [clojure.test :refer [deftest is testing use-fixtures]]
            [glitch.provider :as p]))

;; Reset registry before each test to ensure isolation.
(use-fixtures :each
  (fn [f]
    (p/reset!)
    (f)
    (p/reset!)))

;; ---------------------------------------------------------------------------
;; register + call-provider
;; ---------------------------------------------------------------------------

(deftest register-and-call-test
  (testing "register a mock provider and call it"
    (p/register "mock"
      (fn [{:keys [prompt]}]
        {:response   (str "echo: " prompt)
         :tokens-in  10
         :tokens-out 5}))
    (let [res (p/call-provider "mock" {:prompt "hello"})]
      (is (= "echo: hello" (:response res)))
      (is (= 10 (:tokens-in res)))
      (is (= 5  (:tokens-out res))))))

(deftest call-unknown-provider-test
  (testing "calling an unregistered provider throws"
    (is (thrown-with-msg? clojure.lang.ExceptionInfo #"unknown provider"
          (p/call-provider "nonexistent" {:prompt "x"})))))

;; ---------------------------------------------------------------------------
;; names
;; ---------------------------------------------------------------------------

(deftest names-test
  (testing "names returns sorted list of registered providers"
    (p/register "zulu" identity)
    (p/register "alpha" identity)
    (p/register "mike" identity)
    (is (= ["alpha" "mike" "zulu"] (p/names)))))

;; ---------------------------------------------------------------------------
;; reset!
;; ---------------------------------------------------------------------------

(deftest reset-test
  (testing "reset! clears the registry"
    (p/register "foo" identity)
    (is (= ["foo"] (p/names)))
    (p/reset!)
    (is (= [] (p/names)))))

;; ---------------------------------------------------------------------------
;; call-tiered — escalation
;; ---------------------------------------------------------------------------

(deftest call-tiered-escalation-test
  (testing "first provider fails, second succeeds — escalation works"
    (p/register "bad"
      (fn [_] (throw (ex-info "bad provider" {}))))
    (p/register "good"
      (fn [{:keys [prompt]}]
        {:response   (str "ok: " prompt)
         :tokens-in  1
         :tokens-out 2}))
    (let [res (p/call-tiered {:prompt "test"}
                [{:providers ["bad" "good"]}])]
      (is (= "ok: test" (:response res)))
      (is (= 1 (:tokens-in res)))
      (is (= 2 (:tokens-out res))))))

(deftest call-tiered-cross-tier-test
  (testing "first tier exhausted, second tier succeeds"
    (p/register "fail1"
      (fn [_] (throw (ex-info "nope" {}))))
    (p/register "fail2"
      (fn [_] (throw (ex-info "nope" {}))))
    (p/register "win"
      (fn [_] {:response "victory" :tokens-in 0 :tokens-out 0}))
    (let [res (p/call-tiered {:prompt "x"}
                [{:providers ["fail1" "fail2"]}
                 {:providers ["win"]}])]
      (is (= "victory" (:response res))))))

(deftest call-tiered-model-override-test
  (testing "tier model override is merged into opts"
    (p/register "checker"
      (fn [{:keys [model]}]
        {:response model :tokens-in 0 :tokens-out 0}))
    (let [res (p/call-tiered {:prompt "x" :model "original"}
                [{:providers ["checker"] :model "overridden"}])]
      (is (= "overridden" (:response res))))))

(deftest call-tiered-empty-response-skipped-test
  (testing "provider returning empty response is skipped"
    (p/register "empty"
      (fn [_] {:response "" :tokens-in 0 :tokens-out 0}))
    (p/register "nonempty"
      (fn [_] {:response "content" :tokens-in 0 :tokens-out 0}))
    (let [res (p/call-tiered {:prompt "x"}
                [{:providers ["empty" "nonempty"]}])]
      (is (= "content" (:response res))))))

(deftest call-tiered-nil-response-skipped-test
  (testing "provider returning nil response is skipped"
    (p/register "nilresp"
      (fn [_] {:response nil :tokens-in 0 :tokens-out 0}))
    (p/register "real"
      (fn [_] {:response "real" :tokens-in 0 :tokens-out 0}))
    (let [res (p/call-tiered {:prompt "x"}
                [{:providers ["nilresp" "real"]}])]
      (is (= "real" (:response res))))))

;; ---------------------------------------------------------------------------
;; call-tiered — exhaustion
;; ---------------------------------------------------------------------------

(deftest call-tiered-exhaustion-test
  (testing "all providers fail — throws with exhaustion message"
    (p/register "fail-a"
      (fn [_] (throw (ex-info "a broke" {}))))
    (p/register "fail-b"
      (fn [_] (throw (ex-info "b broke" {}))))
    (is (thrown-with-msg? clojure.lang.ExceptionInfo #"all tiers exhausted"
          (p/call-tiered {:prompt "x"}
            [{:providers ["fail-a"]}
             {:providers ["fail-b"]}])))))

;; ---------------------------------------------------------------------------
;; load-providers — with a temp directory
;; ---------------------------------------------------------------------------

(deftest load-providers-missing-dir-test
  (testing "load-providers gracefully skips missing directories"
    ;; Should not throw
    (p/load-providers "/tmp/glitch-test-nonexistent-dir-12345")))

(deftest load-providers-from-dir-test
  (testing "load-providers loads .clj files that self-register"
    (let [dir  (str "/tmp/glitch-test-providers-" (System/currentTimeMillis))
          file (str dir "/test_prov.clj")]
      (try
        (.mkdirs (java.io.File. dir))
        (spit file
          (str "(require '[glitch.provider :as provider])\n"
               "(provider/register \"test-loaded\"\n"
               "  (fn [opts] {:response \"loaded!\" :tokens-in 0 :tokens-out 0}))\n"))
        (p/load-providers dir)
        (is (some #{"test-loaded"} (p/names)))
        (let [res (p/call-provider "test-loaded" {:prompt "hi"})]
          (is (= "loaded!" (:response res))))
        (finally
          (clojure.java.io/delete-file file true)
          (clojure.java.io/delete-file dir true))))))
