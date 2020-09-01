package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// testing attach and ping for a UE with TNLA.
func AttachUeWithTnla(imsi string, ranUeId int64, ranIpAddr string, wg *sync.WaitGroup, ranPort int) {

	defer wg.Done()

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, ranPort)
	if err != nil {
		fmt.Println("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		fmt.Println("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a UE.
	suciv1, suciv2, err := decodeUeSuci(imsi)
	if err != nil {
		fmt.Println("The test failed when SUCI was created! Error:%s", err)
	}
	err = registrationUE(conn, imsi, ranUeId, suciv2, suciv1, ranIpAddr)
	if err != nil {
		fmt.Println("The test failed when UE tried to attach! Error:%s", err)
	}

	// data plane UE
	// gtpHeader := generateGtpHeader(1)
	// err = pingUE(upfConn, gtpHeader, "60.60.0.1")
	// if err != nil {
	// return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
	//}

	// end sockets.
	conn.Close()
	//upfConn.Close()

	fmt.Println("Thread with imsi:%s worked fine", imsi)
}

// testing authentication for a GNB
func testAttachGnb() error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	// function worked fine.
	return nil
}

// testing multiple GNBs authentication.
func testMultiAttachGnb(numberGnbs int) error {

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	for i := 1; i <= numberGnbs; i++ {

		// multiple names for GNBs.
		nameGNB := "my5gRanTester" + string(i)

		// generate GNB id.
		var aux string
		if i < 16 {
			aux = "00000" + fmt.Sprintf("%x", i)
		} else if i < 256 {
			aux = "0000" + fmt.Sprintf("%x", i)
		} else {
			aux = "000" + fmt.Sprintf("%x", i)
		}

		resu, err := hex.DecodeString(aux)
		if err != nil {
			return fmt.Errorf("error in GNB id for testing multiple GNBs")
		}

		// authentication to a GNB.
		err = registrationGNB(conn, resu, nameGNB)
		if err != nil {
			return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
		}
	}

	// functions worked fine.
	return nil
}

// testing attach and ping for multiple queued UEs.
func testMultiAttachUesInQueue(numberUes int) error {
	const ranIpAddr string = "10.200.200.2"

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	if err != nil {
		return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	}

	// authentication to a GNB.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication and ping to some UEs.
	for i := 1; i <= numberUes; i++ {

		// generating some IMSIs to each UE.
		imsi := generateImsi(i)

		// authentication to a UE.
		suciv1, suciv2, err := decodeUeSuci(imsi)
		if err != nil {
			return fmt.Errorf("The test failed when SUCI was created! Error:%s", err)
		}

		err = registrationUE(conn, imsi, int64(i), suciv2, suciv1, ranIpAddr)
		if err != nil {
			return fmt.Errorf("The test failed when UE tried to attach! Error:%s", err)
		}

		// data plane UE
		ipUe := getSrcPing(i)
		gtpHeader := generateGtpHeader(i)

		err = pingUE(upfConn, gtpHeader, ipUe)
		if err != nil {
			return fmt.Errorf("The test failed when UE tried to use ping! Error:%s", err)
		}
	}

	// end sockets.
	conn.Close()
	upfConn.Close()

	return nil
}

// testing attach and ping for multiple concurrent UEs using 2 GNBs.
func testMultiAttachUesInConcurrencyWithGNBs() error {
	const ranIpAddr string = "10.200.200.2"
	const ran2IpAddr string = "10.200.200.1"

	var wg sync.WaitGroup

	// make N2(RAN connect to AMF)
	conn, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9487)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN connect to UPF)
	//upfConn, err := connectToUpf(ranIpAddr, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// make N2(RAN2 connect to AMF)
	conn2, err := connectToAmf("127.0.0.1", "127.0.0.1", 38412, 9488)
	if err != nil {
		return fmt.Errorf("The test failed when sctp socket 2 tried to connect to AMF! Error:%s", err)
	}

	// make n3(RAN2 connect to UPF)
	//upfConn2, err := connectToUpf(ranIpAddr2, "10.200.200.102", 2152, 2152)
	//if err != nil {
	//	return fmt.Errorf("The test failed when udp socket tried to connect to UPF! Error:%s", err)
	//}

	// authentication to a GNB1.
	err = registrationGNB(conn, []byte("\x00\x01\x02"), "free5gc")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication to a GNB2.
	err = registrationGNB(conn2, []byte("\x00\x01\x01"), "free5gc2")
	if err != nil {
		return fmt.Errorf("The test failed when GNB tried to attach! Error:%s", err)
	}

	// authentication and ping to some concurrent UEs.

	// Launch several goroutines and increment the WaitGroup counter for each.
	wg.Add(1)

	// goroutine.
	go func(wg *sync.WaitGroup) {

		defer wg.Done()

		for i := 1; i <= 5; i++ {

			// generating some IMSIs to each UE.
			imsi := generateImsi(i)

			// authentication to a UE.
			suciv1, suciv2, err := decodeUeSuci(imsi)
			if err != nil {
				fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			}

			err = registrationUE(conn, imsi, int64(i), suciv2, suciv1, ranIpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}
			// thread worked fine.
			fmt.Println("Thread with imsi:%s worked fine", imsi)
		}

	}(&wg)

	// increment the WaitGroup counter.
	wg.Add(1)

	// goroutine.
	go func(wg *sync.WaitGroup) {

		defer wg.Done()

		for i := 6; i <= 10; i++ {
			// generating some IMSIs to each UE.
			imsi := generateImsi(i)

			// authentication to a UE.
			suciv1, suciv2, err := decodeUeSuci(imsi)
			if err != nil {
				fmt.Println("The test failed when SUCI was created! Error:%s in Thread with imsi:%s", err, imsi)
			}

			err = registrationUE(conn2, imsi, int64(i), suciv2, suciv1, ran2IpAddr)
			if err != nil {
				fmt.Println("The test failed when UE tried to attach! Error:%s in Thread with imsi:%s", err, imsi)
			}

			// thread worked fine.
			fmt.Println("Thread with imsi:%s worked fine", imsi)
		}

	}(&wg)

	// wait for multiple goroutines.
	wg.Wait()

	// end sockets.
	conn.Close()
	conn2.Close()
	// upfConn.Close()
	// upfConn2.Close()

	return nil
}

// testing attach and ping for multiple concurrent UEs using TNLAs.
func testMultiAttachUesInConcurrencyWithTNLAs() error {

	var wg sync.WaitGroup

	// authentication and ping to some  concurrent UEs.

	// Launch several goroutines and increment the WaitGroup counter for each.
	wg.Add(1)
	go AttachUeWithTnla("imsi-2089300000001", 1, "10.200.200.2", &wg, 9487)

	wg.Add(1)
	go AttachUeWithTnla("imsi-2089300000002", 2, "10.200.200.2", &wg, 9488)

	wg.Add(1)
	go AttachUeWithTnla("imsi-2089300000003", 3, "10.200.200.2", &wg, 9489)

	// wait for multiple goroutines.
	wg.Wait()

	// function worked fine.
	return nil
}
