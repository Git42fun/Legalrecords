'use strict';

var { Gateway, Wallets, CouchDBWalletStore, X509WalletMixin } = require('fabric-network');
const path = require('path');
const FabricCAServices = require('fabric-ca-client');
const fs = require('fs');

const invoke = require('./invoke.js');

const util = require('util');


const getCCP = async (org) => {
    let ccpPath = null;
    org == 'Org1' ? ccpPath = path.resolve(__dirname, '..', 'config', 'connection-org1.json') : null
    org == 'Org2' ? ccpPath = path.resolve(__dirname, '..', 'config', 'connection-org2.json') : null
    const ccpJSON = fs.readFileSync(ccpPath, 'utf8')
    const ccp = JSON.parse(ccpJSON);
    return ccp
}

const getCaUrl = async (org, ccp) => {
    let caURL = null
    org == 'Org1' ? caURL = ccp.certificateAuthorities['ca.org1.example.com'].url : null
    org == 'Org2' ? caURL = ccp.certificateAuthorities['ca.org2.example.com'].url : null
    return caURL

}

const getWalletPath = async (org) => {
    let walletPath = null
    org == 'Org1' ? walletPath = path.join(process.cwd(), 'org1-wallet') : null
    org == 'Org2' ? walletPath = path.join(process.cwd(), 'org2-wallet') : null
    return walletPath
}


const getAffiliation = async (org) => {
    // Default in ca config file we have only two affiliations, if you want ti use org3 ca, you have to update config file with third affiliation
    //  Here already two Affiliation are there, using i am using "org2.department1" even for org3
    return org == "Org1" ? 'org1.department1' : 'org2.department1'
}

const getRegisteredUser = async (username, userOrg, permissions, isJson) => {
    let ccp = await getCCP(userOrg)

    const caURL = await getCaUrl(userOrg, ccp)
    console.log("ca url is ", caURL)
    const ca = new FabricCAServices(caURL);

    // const couchDBWalletStore = {
    //     url: 'http://admin:password@localhost:5990/', // Replace with your CouchDB URL
    //     walletPath: './couchdb_wallet',   // Replace with your desired wallet path
    //   };
 
    //   const wallet = await Wallets.newCouchDBWallet(couchDBWalletStore);

    const walletPath = await getWalletPath(userOrg)
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    const userIdentity = await wallet.get(username);
    if (userIdentity) {
        console.log(`An identity for the user ${username} already exists in the wallet`);
        var response = {
            success: true,
            message: username + ' enrolled Successfully',
        };
        return response
    }

    // Check to see if we've already enrolled the admin user.
    let adminIdentity = await wallet.get('admin');
    if (!adminIdentity) {
        console.log('An identity for the admin user "admin" does not exist in the wallet');
        await enrollAdmin(userOrg, ccp);
        adminIdentity = await wallet.get('admin');
        console.log("Admin Enrolled Successfully")
    }
    console.log("Admin Enrolled Successfully=====================================", adminIdentity)

    // build a user object for authenticating with the CA
    const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');
    console.log("----------------adminUser-----------------------", adminUser)
    let secret;
    try {
        // Register the user, enroll the user, and import the new identity into the wallet.
     
        if (permissions == "READ-WRITE"){
        secret = await ca.register({ affiliation: 'org1.department1', enrollmentID: username, role: 'client', attrs: [{ name: 'role', value: 'approver', ecert: true }] }, adminUser);
        }
        else{
        secret = await ca.register({ affiliation: await getAffiliation(userOrg), enrollmentID: username, role: 'client', attrs: [{ name: 'role', value: 'client', ecert: true }] }, adminUser);
        }
        console.log(`Secret for the user with username: ${username} -------> ${secret}`)

    } catch (error) {
        return error.message
    }

    const enrollment = await ca.enroll({ enrollmentID: username, enrollmentSecret: secret });
    // const enrollment = await ca.enroll({ enrollmentID: username, enrollmentSecret: secret, attr_reqs: [{ name: 'role', optional: false }] });

    let x509Identity = {
        credentials: {
            certificate: enrollment.certificate,
            privateKey: enrollment.key.toBytes(),
        },
        mspId: `${userOrg}MSP`,
        type: 'X.509',
    };
    await wallet.put(username, x509Identity);

    console.log(`Successfully registered and enrolled admin user ${username} and imported it into the wallet`);

    var response = {
        success: true,
        message: username + ' enrolled Successfully',
    };
    return response
}

const isUserRegistered = async (username, userOrg) => {
    const walletPath = await getWalletPath(userOrg)
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    const userIdentity = await wallet.get(username);
    if (userIdentity) {
        console.log(`An identity for the user ${username} exists in the wallet`);
        return true
    }
    return false
}


const getCaInfo = async (org, ccp) => {
    let caInfo = null
    org == 'Org1' ? caInfo = ccp.certificateAuthorities['ca.org1.example.com'] : null
    org == 'Org2' ? caInfo = ccp.certificateAuthorities['ca.org2.example.com'] : null
    return caInfo
}

const getOrgMSP = (org) => {
    let orgMSP = null
    org == 'Org1' ? orgMSP = 'Org1MSP' : null
    org == 'Org2' ? orgMSP = 'Org2MSP' : null
    return orgMSP

}

const enrollAdmin = async (org, ccp) => {
    console.log('calling enroll Admin method')
    try {
        const caInfo = await getCaInfo(org, ccp) //ccp.certificateAuthorities['ca.org1.example.com'];
        const caTLSCACerts = caInfo.tlsCACerts.pem;
        const ca = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);


        // const couchDBWalletStore = {
        //     url: 'http://admin:password@localhost:5990/', // Replace with your CouchDB URL
        //     walletPath: './couchdb_wallet',   // Replace with your desired wallet path
        //   };
    
        //   const wallet = await Wallets.newCouchDBWallet(couchDBWalletStore);

        // Create a new file system based wallet for managing identities.
        const walletPath = await getWalletPath(org) //path.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the admin user.
        const identity = await wallet.get('admin');
        if (identity) {
            console.log('An identity for the admin user "admin" already exists in the wallet');
            return;
        }

        // Enroll the admin user, and import the new identity into the wallet.
        const enrollment = await ca.enroll({ enrollmentID: 'admin', enrollmentSecret: 'adminpw' });
        console.log("Enrollment object is : ", enrollment)
        let x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: `${org}MSP`,
            type: 'X.509',
        };

        // const identityLabel = 'admin'; // Specify the identity label
        // const identity1 = X509WalletMixin.createIdentity('Org1MSP', certificate, privateKey);

        // await wallet.put(identityLabel, identity1);

        await wallet.put('admin', x509Identity);
        console.log('Successfully enrolled admin user "admin" and imported it into the wallet');
        return
    } catch (error) {
        console.error(`Failed to enroll admin user "admin": ${error}`);
    }
}

const registerAndGerSecret = async (username, userOrg, name, password, userType, permissions) => {
    let ccp = await getCCP(userOrg)

    const caURL = await getCaUrl(userOrg, ccp)
    const ca = new FabricCAServices(caURL);

    const walletPath = await getWalletPath(userOrg)
    const wallet = await Wallets.newFileSystemWallet(walletPath);
    console.log(`Wallet path: ${walletPath}`);

    const userIdentity = await wallet.get(username);
    if (userIdentity) {
        console.log(`An identity for the user ${username} already exists in the wallet`);
        var response = {
            success: true,
            message: username + ' enrolled Successfully',
        };
        return response
    }

    // Check to see if we've already enrolled the admin user.
    let adminIdentity = await wallet.get('admin');
    if (!adminIdentity) {
        console.log('An identity for the admin user "admin" does not exist in the wallet');
        await enrollAdmin(userOrg, ccp);
        adminIdentity = await wallet.get('admin');
        console.log("Admin Enrolled Successfully")
    }

    // build a user object for authenticating with the CA
    const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, 'admin');
    let secret;
    try {
        // Register the user, enroll the user, and import the new identity into the wallet.

        if (permissions == "READ-WRITE"){
            secret = await ca.register({ affiliation: 'org1.department1', enrollmentID: username, role: 'client', attrs: [{ name: 'role', value: 'approver', ecert: true }] }, adminUser);
            }
            else{
            secret = await ca.register({ affiliation: await getAffiliation(userOrg), enrollmentID: username, role: 'client', attrs: [{ name: 'role', value: 'client', ecert: true }] }, adminUser);
            }
        // const secret = await ca.register({ affiliation: 'org1.department1', enrollmentID: username, role: 'client', attrs: [{ name: 'role', value: 'approver', ecert: true }] }, adminUser);
        const enrollment = await ca.enroll({
            enrollmentID: username,
            enrollmentSecret: secret
        });
        let orgMSPId = getOrgMSP(userOrg)
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: orgMSPId,
            type: 'X.509',
        };
        await wallet.put(username, x509Identity);

        // Add the user to the User struct using invokeTransaction
        let args = [username, name, password, userType, permissions];
        let message = await invoke.invokeTransaction('mychannel', 'fabcar', 'CreateUser', args, username, userOrg, null);
        
        const response_payload = {
            result: message,
            error: null,
            errorData: null
        }
        res.send(response_payload);
    } catch (error) {
        return error.message
    }

    var response = {
        success: true,
        message: username + ' enrolled Successfully',
        secret: secret
    };
    return response

}

exports.getRegisteredUser = getRegisteredUser
exports.getCCP = getCCP
exports.getWalletPath = getWalletPath
module.exports = {
    getCCP: getCCP,
    getWalletPath: getWalletPath,
    getRegisteredUser: getRegisteredUser,
    isUserRegistered: isUserRegistered,
    registerAndGerSecret: registerAndGerSecret

}
