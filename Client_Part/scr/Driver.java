import java.io.FileReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.PublicKey;

import javax.crypto.BadPaddingException;
import javax.crypto.Cipher;
import javax.crypto.IllegalBlockSizeException;
import javax.crypto.NoSuchPaddingException;

import org.json.simple.JSONObject; 
import org.json.simple.parser.*; 


public class Driver {
	
	public static String resume_path = "/resume.json";
	public static int numOfZero = 3;
	
	
	public static void main(String[] args) {
		readResume();
	}

	/**
	 * Read the resume json file
	 */
	private static void readResume() {
		String cwd = System.getProperty("user.dir");
		
		
		try {
			Object obj = new JSONParser().parse(new FileReader(cwd.concat(resume_path)));
			JSONObject resumeJSON = (JSONObject) obj;
			int nonce = 0;

			KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA");
			kpg.initialize(2048);
			KeyPair kp = kpg.generateKeyPair();
			
			PublicKey pub = kp.getPublic();
			PrivateKey prv = kp.getPrivate();
			
			while(true) {
				MessageDigest digest = MessageDigest.getInstance("SHA-256");
				byte[] hash = digest.digest(resumeJSON.toString().concat(Integer.toString(nonce)).getBytes(StandardCharsets.UTF_8));
				boolean flag = true;
				for(int i = 0; i < numOfZero; i++) {
					if(hash[i] != 0) {
						flag = false;
						
					}
				}
				if(flag) {
					sign(prv, hash);
					break;
				}
				nonce++;
			}
			
			//send the pub key and hash to server
			
			
			
			
			
			
		} catch (IOException e) {
			System.err.println("file does not exist");
		} catch (NoSuchAlgorithmException e) {
			System.err.println("encryption error, this computer does not have such encryption method");
		} catch (ParseException e) {
			System.err.println("error during paring json file, please check the format and content");
		}
	}
	
	public static String byteArrayToHex(byte[] a) {
		   StringBuilder sb = new StringBuilder(a.length * 2);
		   for(byte b: a)
		      sb.append(String.format("%02x", b));
		   return sb.toString();
		}
	
	public static String sign(PrivateKey pry, byte[] hash) {
		
		byte[] obuf = new byte[] {};
		try {
			Cipher cipher = Cipher.getInstance("RSA/ECB/PKCS1Padding");
			cipher.init(Cipher.DECRYPT_MODE, pry);
			obuf = cipher.update(hash, 0, hash.length);
			obuf = cipher.doFinal();
			
		} catch (NoSuchAlgorithmException | NoSuchPaddingException | InvalidKeyException | IllegalBlockSizeException | BadPaddingException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
		return byteArrayToHex(obuf);
	
	}

}
