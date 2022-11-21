




Achtung: die in dem Diagramm dargestellten Kommunikationsbeschreibungen entsprechen nicht der Methoden-Benennung im Code!

# Beschreibung der Microservices:

## Customer-Service:

Beschreibung:
- Microservice zur Verwaltung der Kundendaten.

Datenhaltung:
- Kundennummer
- Name
- Adresse

Abhängigkeiten:
- Keine



## Order-Service:
Beschreibung:
Microservice zur Verwaltung der Bestellungen von Kunden.  Nach Aufgabe einer Bestellung wartet der Payment-Service auf das Bezahlen der Rechnungssumme. Meldet der Payment-Service, dass die Bestellung bezahlt wurde, wird die Bestellung durch den Shipment-Service versandt. Meldet der Shipment-Service, dass die Bestellung Versand wurde, wird die Bestellung als abgeschlossen markiert.
Meldet der Shipment-Service eine Rücksendung,  wird der zurückgegeben Artikel aus der Bestellung gelöscht und die Kosten für den Artikel über den Payment-Service zurückerstattet.
Wird die Order vom Kunden storniert, wird der Shipment-Service und der Payment-Service vom Order-Service darüber informiert.
Um Redundanz zu vermeiden, speichert der Order-Service keine Kundendaten ab, sondern holt sich diese jedes Mal vom Customer-Service.


Datenhaltung:
- Bestellnummer
- Kundennummer
- Bestelle Artikel mit Stückzahl
- Gesamtkosten
- Bezahlt (Ja/Nein)
- Versendet (Ja/Nein)
- Storniert (Ja/Nein)


Abhängigkeiten:
- Order-Service -> Customer-Service:
    - Synchrone Kommunikation zum Überprüfen, ob der Kunde der die Bestellung aufgegeben hat existiert. Synchron, da Antwort benötigt wird.
- Order-Service -> Catalog-Service:
    - Synchrone Kommunikation zum Überprüfen, ob bestellter Artikel existiert (bzw. verkauft wird). Synchron, da Antwort benötigt wird.
- Order-Service -> Payment-Service:
    - Asynchrone Kommunikation, um nach Abschluss einer Bestellung den Bezahlungsprozess einzuleiten. Asynchron, da Bezahlprozess nur angestoßen und keine direkte Rückmeldung benötigt wird.
    - Asynchrone Kommunikation, um nach Eingang einer Retoure den Retourenbetrag zurückzuzahlen. Asynchron, da Auszahlung nur angestoßen und keine direkte Rückmeldung benötigt wird.
    - Asynchrone Kommunikation, um nach Stornierung der Bestellung den bereits bezahlten Betrag zurückzuzahlen. Asynchron, da Auszahlung nur angestoßen und keine direkte Rückmeldung benötigt wird.
- Order-Service -> Shipment-Service:
    - Asynchrone Kommunikation, um nach erfolgreicher Bezahlung der Bestellung den Versandprozess einzuleiten. Asynchron, da Versandprozess nur angestoßen und keine direkte Rückmeldung benötigt wird.
    - Asynchrone Kommunikation, um nach Stornierung der Bestellung dem Shipment-Service Bescheid zu geben, dass die Bestellung storniert wurde und er reservierte Artikel freigeben kann. Asynchron, da Stornierungsprozess nur angestoßen und keine direkte Rückmeldung benötigt wird.



## Payment-Service:
Beschreibung:
Microservice zur Verwaltung der Zahlungen. Biete die Möglichkeit im Order-Service aufgegebene Bestellungen zu bezahlen. Speichert alle offenen und abgewickelten Bezahlungen.
Nach der Aufgabe einer Bestellung wird Bestellnummer und Rechnungssumme dem Payment-Service übergeben. Trift eine Zahlung (oder mehrere kleine) mit Angabe der Bestellnummer in Höhe der Rechnungssumme ein, so wird der Bezahlvorgang abgeschlossen und eine Nachricht an den Shipment-Service geschickt.
Muss eine Rücksendung zurückbezahlt werden, wird das zur Bestellnummer gehörende Payment geladen und die Rechnungssumme sowie der vom Kunden bezahlte Betrag um den Retourenbetrag vermindert. Anschließend wird der Retourenbetrag dem Kunden zurückbezahlt.
Wird ein Payment storniert, wird der bereits vom Kunde bezahlte Betrag zurückbezahlt und die Rechnungssumme auf 0 gesetzt.


Datenhaltung:
- Bestellnummer
- Rechnungssumme
- Bereits bezahlter Betrag
- Storniert (Ja/Nein)

Abhängigkeiten:
- Payment-Service -> Order-Service:
    - Asynchrone Kommunikation, um nach Bezahlung der gesamten Bestellsumme zu melden, dass Bestellung bezahlt wurde.

- Payment-Provider API:
    - Synchrone Kommunikation, um Payment zu bezahlen. Synchron, da Schnittstelle zu Drittanbieter und weil Payment-Service den restlichen zu zahlenden Betrag zurückliefert.


## Shipment-Service:
Beschreibung:
Microservice zur Verwaltung von Versendungen und Retouren. Wurde eine Zahlung bestätigt, bucht der Shipment-Service mithilfe des Stock-Service alle Produkte aus dem Lager aus und der Shipment-Service versendet die Produkte. Falls Artikel nicht verfügbar sind, werden sie, sobald sie geliefert wurden, an den Shipment-Service gesendet.
Kommen Rücksendungen im Lager an, können diese entweder sofort mit neuen Artikeln aus dem Lager mit Hilfe des Stock-Service ersetzt werden oder es wird dem Order-Service kommuniziert, dass eine Rückerstattung des Teilbetrags an den Kunden gesendet werden soll.
Falls eine Bestellung storniert wird, werden alle schon ausgebuchten Artikel zurück ins Lager gegeben und die Reservierung von noch nicht vorhanden Artikeln wird gelöscht.


Datenhaltung:
- Sendungsnummer/Bestellnummer
- Artikel
- Status
- Adresse

Abhängigkeiten:
- Shipment-Service -> Order-Service:
    - Asynchrone Kommunikation, um nach Versand der Bestellung zu melden, dass Bestellung Versand wurde. Asynchron, da die Meldung nur zur Info ist.

- Shipment-Service -> Stock-Service:
    - Synchrone Kommunikation zum Abfragen, ob alle Artikel einer Bestellung lagernd und versandbereit sind. Synchron, da die Antwort relevant ist.
    - Asynchrone Kommunikation zum Freigeben der Artikel, falls eine Bestellung storniert wird. Asynchron, da keine Antwort benötigt wird.
    
- Shipment API:
    -  Synchrone Kommunikation, um die Produkte an den Kunden zu senden. Synchron, da es eine Schnittstelle nach außen ist.



## Stock-Service:
Beschreibung:
Microservice zur Verwaltung der Lagerhaltung. Speichert die Stückzahl jedes Artikels und bietet die Möglichkeit Artikel zu reservieren. Trifft eine neue Lieferung ein, können "Supplier" diese dem Stock-Service melden und dem Lager hinzufügen.

Datenhaltung:
- Artikelnummer
- Stückzahl
- Anzahl reservierter Artikel mit Sendenummer


Abhängigkeiten:
- Stock-Service -> Shipment-Service:
-   Asynchrone Kommunikation, um neu eingetroffene Produkte zu den benötigten Warenauslieferungen zu bringen. Asynchron, da Stock-Service nicht unbedingt eine Antwort benötigt.

- Stock-Service -> Supplier-Service:
-   Asynchrone Kommunikation, um Artikel nachzubestellen. Asynchron, da Stock-Service nicht unbedingt eine Antwort benötigt und die Bestätigung, dass der Artikel beim Zulieferer bestellt wurde, relativ lange dauert.


## Supplier-Service:
Beschreibung:
Microservice zur Verwaltung der Lieferungen von Zulieferern. Meldet dem Stock-Service, wenn eine neue Lieferung hinzugefügt wurde. Es werden alle Wareneingänge in einer Datenbank gespeichert.
Es wird ebenfalls die Möglichkeit geboten neue Artikel bei den Zulieferern zu bestellen.

Datenhaltung:
- Supplier Name
- Artikelnummer
- Stückzahl gelieferter Artikel
- Liefernummer

Abhängigkeiten:
- Supplier-Service -> Stock-Service:
    - Asynchrone Kommunikation, um nach Wareneingang zu melden, dass neue oder ein Nachschub an Artikel eingetroffen ist. Asynchron, da keine Antwort erwartet wird.

- Supplier API:
    - Synchrone Kommunikation, um neue Artikel beim Zulieferer zu bestellen. Synchron, da es eine Schnittstelle nach außen ist.



## Catalog-Service:
Beschreibung:
Microservice zur Verwaltung der Artikel. Kann angefragt werden für Informationen über die Produkte. Fragt synchron die Verfügbarkeit von Artikeln bei dem Stock-Service an.

Datenhaltung:
- Artikelnummer
- Artikelname
- Beschreibung
- Preis

Abhängigkeiten:
- Catalog-Service -> Stock-Service:
    - Synchrone Kommunikation, um die Verfügbarkeit von Artikeln zu überprüfen. Bei einer Just-in-Time Lagerhaltung ändern sich die Zustände des Lagers mit großer Wahrscheinlichkeit häufiger als Kunden etwas bestellen (synchron ist weniger Kommunikation nötig als asynchron und gecached).


# How to use: 

cmd commands:
- make run1
- make run2
- make run3
- make run4
- make run5
- make run
