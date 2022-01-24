# usage: make run1
# docker-compose -f A -f B "merges" A & B

.PHONY: test run1 run2 run3 run4 run5 run

test:
	@echo "------------------------------------------------------"
	@echo "-- TEST ----------------------------------------------"
	@echo "------------------------------------------------------"
	docker-compose -f docker-compose.yaml -f docker-compose.test.yaml up --abort-on-container-exit

run1:
	@echo "------------------------------------------------------"
	@echo "-- SZENARIO 1 ----------------------------------------"
	@echo "------------------------------------------------------"
	docker-compose -f docker-compose.yaml -f docker-compose.S1.yaml up --abort-on-container-exit

run2:
	@echo "------------------------------------------------------"
	@echo "-- SZENARIO 2 ----------------------------------------"
	@echo "------------------------------------------------------"
		docker-compose -f docker-compose.yaml -f docker-compose.S2.yaml up --abort-on-container-exit

run3:
	@echo "------------------------------------------------------"
	@echo "-- SZENARIO 3 ----------------------------------------"
	@echo "------------------------------------------------------"
	docker-compose -f docker-compose.yaml -f docker-compose.S3.yaml up --abort-on-container-exit

run4:
	@echo "------------------------------------------------------"
	@echo "-- SZENARIO 4 ----------------------------------------"
	@echo "------------------------------------------------------"
	docker-compose -f docker-compose.yaml -f docker-compose.S4.yaml up --abort-on-container-exit

run5:
	@echo "------------------------------------------------------"
	@echo "-- SZENARIO 5 ----------------------------------------"
	@echo "------------------------------------------------------"
	docker-compose -f docker-compose.yaml -f docker-compose.S5.yaml up --abort-on-container-exit

run: run1 run2 run3 run4 run5