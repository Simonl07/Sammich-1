import java.io.FileReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.KeyFactory;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.PublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.Base64;

import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;

import org.json.simple.JSONObject;
import org.json.simple.parser.JSONParser;
import org.json.simple.parser.ParseException;



public class Driver {

	public static String resume_path = "/resume.json";
	private static String decrpytGlobal = "";
	private static PrivateKey prvGlobal;


	public static void main(String[] args) throws InvalidKeySpecException {
		JSONObject resume = readResume();
		
		sendRequestion(resume);
		
		
		
		
		
	}

	private static void sendRequestion(JSONObject resume) {

		JSONObject newApplcation = new JSONObject();
		
		JSONObject newID = new JSONObject();
		JSONObject resumeId = (JSONObject)resume.get("Indentity");
		newID.put("Name", (String)resumeId.get("Name"));
		newID.put("Id", (int) resumeId.get("Id"));
		newID.put("Address", (String) resumeId.get("Address"));
		newID.put("Email", (String) resumeId.get("Email"));
		newID.put("Phone", (String) resumeId.get("Phone"));
		newApplcation.put("Indentity", newID);
		
		JSONObject newMent = new JSONObject();
		JSONObject resumeMent = (JSONObject)resume.get("Ments");
		newMent.put("Skills", (String[])resumeId.get("Skills"));
		newMent.put("Education", (String[])resumeId.get("Education"));
		newMent.put("Experiences", (String[])resumeId.get("Experiences"));
		newApplcation.put("Ments", newMent);
		
//		newApplcation.put("PubKey", value)
		
		
		
		
		
		
		
		
	}

	private static PublicKey StringToPubKey(String pub_str) {
		byte[] byte_pubkey  = Base64.getDecoder().decode(pub_str);
		PublicKey factory;
//		PublicKey public_key =  factory.generatePublic(new X509EncodedKeySpec(byte_pubkey));
		
		
		
		return null;
		
//		byte[] byte_pubkey = public_key.getEncoded();
	}
	
	
	
	
	/**
	 * Read, hashing resume and encrypt hashing
	 * @return 
	 * @throws InvalidKeySpecException 
	 */
	private static JSONObject readResume() throws InvalidKeySpecException {
		String cwd = System.getProperty("user.dir");
		JSONObject resumeJSON = null;

		try {
			Object obj = new JSONParser().parse(new FileReader(cwd.concat(resume_path)));
			resumeJSON = (JSONObject) obj;
			int nonce = 0;

			KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
			kpg.initialize(2048);
			KeyPair kp = kpg.generateKeyPair();

			PublicKey pub = kp.getPublic();
			String hex = byteArrayToHex(pub.getEncoded());
			System.out.println(hex);
			
			byte[] b = hexToByteArray(hex);
			X509EncodedKeySpec spec =
				      new X509EncodedKeySpec(b);
				    KeyFactory kf = KeyFactory.getInstance("RSA");
		PublicKey p = kf.generatePublic(spec);
			System.out.println(byteArrayToHex(p.getEncoded()));
			System.out.println(pub.getFormat());
			PrivateKey prv = kp.getPrivate();
			pubKeyToString(pub);
			
			String signature = "";
			String hexString = "";
			
			while(true) {
				//hashing with SHA-256
				MessageDigest digest = MessageDigest.getInstance("SHA-256");
				byte[] hash = digest.digest(resumeJSON.toString().concat(Integer.toString(nonce)).getBytes(StandardCharsets.UTF_8));
				hexString = byteArrayToHex(hash).trim();

				if(hexString.startsWith("0000")) {
					signature = sign(prv, hash).trim();
					break;
				}
				nonce++;
			}

			//send the pub key and hash to server
			byte[] decry = decrypt(signature, pub);

			//decrypted hashing
			String deccrypt = byteArrayToHex(decry).trim();
			decrpytGlobal = deccrypt;

		} catch (IOException e) {
			System.err.println("file does not exist");
		} catch (NoSuchAlgorithmException e) {
			System.err.println("encryption error, this computer does not have such encryption method");
		} catch (ParseException e) {
			System.err.println("error during paring json file, please check the format and content");
		}
		
		return resumeJSON;
	}

	private static void pubKeyToString(PublicKey pub) {
		// TODO Auto-generated method stub
		
		
		
		
		
		
		
		
		
	}

	/**
	 * Convert string to byte array
	 * @param str string
	 * @return byte array of the string
	 */
	public static byte[] hexToByteArray(String str) {
		int len = str.length();
		byte[] data = new byte[len / 2];
		for (int i = 0; i < len; i += 2) {
			data[i / 2] = (byte) ((Character.digit(str.charAt(i), 16) << 4)
					+ Character.digit(str.charAt(i+1), 16));
		}
		return data;
	}

	/**
	 * Convert byte array to hex string
	 * @param array byte array 
	 * @return hex string of the byte array
	 */
	public static String byteArrayToHex(byte[] array) {
		StringBuilder sb = new StringBuilder(array.length * 2);
		for(byte b: array)
			sb.append(String.format("%02x", b));
		return sb.toString();
	}

	/**
	 * Sign the hashing with private key
	 * @param pry private key
	 * @param hash byte array after encrypting
	 * @return String representation of the signed application
	 */
	public static String sign(PrivateKey pry, byte[] hash) {

		byte[] obuf;
		try {
			Cipher cipher = Cipher.getInstance("RSA/ECB/PKCS1Padding");
			cipher.init(Cipher.ENCRYPT_MODE, pry);
			obuf = cipher.update(hash, 0, hash.length);
			obuf = cipher.doFinal();
			return byteArrayToHex(obuf);
		} catch (NoSuchAlgorithmException | NoSuchPaddingException | InvalidKeyException | IllegalBlockSizeException | BadPaddingException e) {
			e.printStackTrace();
			return null;
		}
	}

	/**
	 * Drcryption using public key
	 * @param signature String with applicants signature
	 * @param pub public key
	 * @return byte array of the drcrypted hashing
	 */
	public static byte[] decrypt(String signature, PublicKey pub) {

		byte[] obuf;
		try {

			Cipher cipher = Cipher.getInstance("RSA/ECB/PKCS1Padding");
			cipher.init(Cipher.DECRYPT_MODE, pub);
			byte[] hexString = hexToByteArray(signature);
			obuf = cipher.update(hexString, 0, hexString.length);
			obuf = cipher.doFinal();
			return obuf;
		} catch (NoSuchAlgorithmException | NoSuchPaddingException | InvalidKeyException | IllegalBlockSizeException | BadPaddingException e) {
			e.printStackTrace();
			return null;
		}



	}

}
