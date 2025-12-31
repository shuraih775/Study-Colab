package com.studycollab.file_service.config;

import io.minio.*;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "minio")
@Data
@Slf4j
public class MinioConfig {

    private String internalEndpoint;
    private String publicEndpoint;
    private String minioRegion;
    private String accessKey;
    private String secretKey;
    private String bucket;

    @Bean
    @Qualifier("internalMinio")
    public MinioClient internalMinioClient() {
        MinioClient client = MinioClient.builder()
                .endpoint(internalEndpoint)
                .region(minioRegion)
                .credentials(accessKey, secretKey)
                .build();

        try {
            boolean exists = client.bucketExists(
                    BucketExistsArgs.builder().bucket(bucket).build()
            );

            if (!exists) {
                client.makeBucket(
                        MakeBucketArgs.builder().bucket(bucket).build()
                );
                log.info("MinIO bucket '{}' created", bucket);
            }
        } catch (Exception e) {
            throw new IllegalStateException("MinIO internal init failed", e);
        }

        return client;
    }

    @Bean
    @Qualifier("publicMinio")
    public MinioClient publicMinioClient() {
        return MinioClient.builder()
                .endpoint(publicEndpoint)
                .region(minioRegion)
                .credentials(accessKey, secretKey)
                .build();
    }
}
